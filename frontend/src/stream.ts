// stream.ts
import { ChatCompletionMessageParam } from "openai/src/resources/chat/completions";

export function newStream(baseUrl = "/chat-stream") {
  return new Stream(baseUrl);
}

export class Stream {
  eventSource: EventSource | null = null;
  url: string;
  fullResponse: string = "";
  onChunk: (chunk: string, isDone: boolean) => void = () => {};
  onDone: (finalContent: string) => void = () => {};
  currentThreadId: string | null = null;
  private connectionState: "connected" | "disconnected" | "connecting" =
    "disconnected";

  constructor(url: string) {
    this.url = url;
  }

  async init(threadId?: string) {
    // Close any existing connection
    this.close();
    console.log({ threadId });
    this.connectionState = "connecting";

    // Append conversation ID to URL if provided
    const connectionUrl = threadId
      ? `${this.url}?thread_id=${encodeURIComponent(threadId)}`
      : this.url;

    console.log(`Creating new EventSource connection to ${connectionUrl}`);

    // Create new EventSource
    this.eventSource = new EventSource(connectionUrl);
    this.currentThreadId = threadId || null;

    // Set up event handlers
    this.eventSource.onopen = () => {
      console.log("EventSource connection opened successfully");
      this.connectionState = "connected";
    };

    this.eventSource.addEventListener("open", (event) => {
      console.log("Connected to stream:", event);
      this.connectionState = "connected";
    });

    // Improved message event handler for SSE
    this.eventSource.addEventListener("message", (event) => {
      try {
        // First, check if the data is already valid JSON
        let parsedData;

        try {
          parsedData = JSON.parse(event.data);
        } catch (jsonError) {
          // If JSON parsing fails, the data might be a string or have extra formatting
          console.warn(
            "Initial JSON parse failed, trying to clean the data:",
            jsonError,
          );

          // Some SSE implementations might add quotes or escape characters to the data
          // Try to clean it up
          const cleanedData = event.data
            .replace(/^"/, "") // Remove leading quote if present
            .replace(/"$/, "") // Remove trailing quote if present
            .replace(/\\"/g, '"') // Replace escaped quotes with actual quotes
            .replace(/\\\\/g, "\\"); // Replace double backslashes with single backslashes

          try {
            parsedData = JSON.parse(cleanedData);
          } catch (cleanedJsonError) {
            // If it's still not valid JSON, treat the entire data as a content string
            console.warn(
              "Cleaned JSON parse failed, treating as plain text:",
              cleanedJsonError,
            );
            parsedData = { content: event.data, isDone: false };
          }
        }

        // Process the data regardless of how we parsed it
        if (parsedData.content !== undefined) {
          this.fullResponse += parsedData.content;
        }

        // Determine if the stream is done - check both possible flags for compatibility
        const isDone = parsedData.isDone === true;

        // Call the callback with the new chunk and completion status
        if (parsedData.content === "torgonestjolie") return; // we use this just to open the stream
        this.onChunk(parsedData.content || "", isDone);

        if (isDone) {
          console.log("Stream completed. Full response:", this.fullResponse);
          this.onDone(this.fullResponse);
          // Note: We don't close the connection here - it should remain open for future messages
        }
      } catch (error) {
        console.error(
          "Error processing event data:",
          error,
          "Raw data:",
          event.data,
        );

        // Fallback: still try to pass something to the UI
        this.onChunk(event.data, false);
      }
    });

    this.eventSource.addEventListener("done", (event) => {
      console.log("Stream done event:", event.data);
      // We don't close the connection here either - it's a persistent connection
    });

    this.eventSource.onerror = (event) => {
      console.error("EventSource error:", event);
      this.connectionState = "disconnected";

      // Implement a reconnection strategy
      setTimeout(() => {
        console.log("Attempting to reconnect...");
        this.init(this.currentThreadId || undefined);
      }, 3000);
    };
  }

  onMessage(callback: (content: string, isDone: boolean) => void) {
    this.onChunk = callback;
  }

  isConnected(): boolean {
    console.log("Checking connection state:", {
      readystate: this.eventSource?.readyState,
    });
    return this.connectionState === "connected" && this.eventSource !== null;
  }

  close() {
    if (this.eventSource) {
      console.log("Closing EventSource connection");
      this.eventSource.close();
      this.eventSource = null;
      this.connectionState = "disconnected";
      this.fullResponse = ""; // Reset the response buffer
    }
  }
}
