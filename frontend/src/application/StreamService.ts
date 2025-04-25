// StreamService.ts

import { Stream, newStream } from "../stream";
import { ChatService } from "./ChatService";
import { DexieThreadRepository } from "../infrastructure/persistence/dexieThreadRepository";
import { Message } from "../domain";

class StreamService {
  static instance: StreamService | null = null;
  private chatService: ChatService;
  threadId: string;
  stream: Stream;
  private onChunkCallback: (chunk: string) => void;

  private constructor(threadId: string, onChunk: (chunk: string) => void) {
    this.stream = newStream();
    this.threadId = threadId;
    this.onChunkCallback = onChunk;
    const threadRepository = new DexieThreadRepository();
    this.chatService = new ChatService(threadRepository, this); // Pass StreamService to ChatService
    this.initStream(threadId);
  }

  static getInstance(
    threadId: string,
    onChunk: (chunk: string) => void,
  ): StreamService {
    if (
      !StreamService.instance ||
      StreamService.instance.threadId !== threadId
    ) {
      StreamService.instance = new StreamService(threadId, onChunk);
    } else {
      // Update the callback if the thread is the same but callback changed
      StreamService.instance.onChunkCallback = onChunk;
    }
    return StreamService.instance;
  }

  currentMessage = "";

  initStream(threadId: string) {
    this.stream.init(threadId);
    this.stream.onDone = (finalContent: string) => {
      console.log("streaming done:", finalContent);
      this.chatService.handleMessageChunk({
        finalContent: finalContent || "",
        threadId: this.threadId,
      });
      // Don't close the stream here, as it should remain open for future messages
    };
    this.stream.onMessage((chunk) => {
      const str = chunk || "";
      this.onMessageChunk(str);
      this.onChunkCallback(str);
    });
  }

  onMessageChunk = async (chunk: string) => {
    if (!this.threadId) {
      console.error("Thread ID is not set");
      return;
    }
    this.chatService.handleMessageChunk({ chunk, threadId: this.threadId });
  };

  isStreamConnected(): boolean {
    return this.stream.isConnected();
  }

  reconnectIfNeeded(): void {
    if (!this.stream.isConnected()) {
      console.log("Stream disconnected, reconnecting...");
      this.initStream(this.threadId);
    }
  }

  closeStream(): void {
    this.stream.close();
  }
}

export const createStreamService = (
  threadId: string,
  onChunk: (chunk: string) => void,
): StreamService => {
  return StreamService.getInstance(threadId, onChunk);
};

export const getStreamService = (): StreamService | null => {
  return StreamService.instance;
};
