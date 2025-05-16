import { generateUUID } from "../utils";
import {
  Thread,
  Message,
  MessageRole,
  Attachment,
  ThreadRepository,
  messageStatusMap,
  messageRoleMap,
  ModelParams,
} from "../domain";
import { createMessage } from "../db";
import api from "../api";
import { getStreamService } from "./StreamService";
import { ChatCompletionMessageParam } from "openai/src/resources/chat/completions";

export class ChatService {
  private repository: ThreadRepository;
  private streamService: any; // Using any to avoid circular dependency

  constructor(repository: ThreadRepository, streamService?: any) {
    this.repository = repository;
    this.streamService = streamService;
  }

  async getModelParams(threadId: string): Promise<ModelParams | undefined> {
    const messages = await this.repository.getMessagesByThreadId(threadId);
    const userMessages = messages.filter((m) => m.role === messageRoleMap.USER);
    if (userMessages.length) {
      return userMessages[userMessages.length - 1].modelParams;
    }

    return undefined;
  }

  async createThread(title?: string): Promise<string> {
    const thread = new Thread({
      id: generateUUID(),
      title: title || "New Thread",
      createdAt: new Date(),
    });

    await this.repository.saveThread(thread);
    return thread.id;
  }

  async getThread(id: string): Promise<Thread | undefined> {
    return this.repository.getThreadById(id);
  }

  handleMessageChunk = async (params: {
    threadId: string;
    chunk?: string;
    finalContent?: string;
  }) => {
    const messages = await this.repository.getMessagesByThreadId(
      params.threadId,
    );
    const assistantMessage = messages.filter(
      (m) => m.role === "assistant" && m.status === messageStatusMap.STREAMING,
    );

    const lastAssistantMessage = assistantMessage[assistantMessage.length - 1];

    if (!lastAssistantMessage) {
      console.error("Assistant message not found");
      return;
    }

    if (params.finalContent) {
      lastAssistantMessage.status = messageStatusMap.DONE;
    } else {
      lastAssistantMessage.content += params.chunk || "";
    }

    await this.saveMessage(lastAssistantMessage);
  };

  async listThreads(): Promise<Thread[]> {
    return this.repository.listThreads();
  }

  async renameThread(id: string, newTitle: string): Promise<void> {
    console.log(`Renaming thread with id ${id} to ${newTitle}`);
    const thread = await this.getThread(id);
    if (!thread) throw new Error("Thread not found");
    thread.rename(newTitle);
    await this.repository.saveThread(thread);
  }

  async getMessages(threadId: string): Promise<Message[]> {
    return this.repository.getMessagesByThreadId(threadId);
  }

  async getAttachements(threadId: string): Promise<Attachment[]> {
    const messages = await this.repository.getMessagesByThreadId(threadId);
    return messages.flatMap((m) => m.attachments || []);
  }

  async generateThreadName(threadId: string, content: string) {
    console.log("Generating thread name");
    const prompt = `
              I want you to summarize the message below into 1 short sentence so it can serve as the title of the conversation the message is opening.
              Your response will be directly serving as the title, so please just respond with the title and nothing else. The content might be in another language than english, so please be careful with the language you use.
              message:
              ${content}
          `;
    const data = await this.getCompletion([{ role: "user", content: prompt }]);
    console.log({ name: data.content });

    await this.renameThread(threadId, data.content);
  }

  async getCompletion(messages: ChatCompletionMessageParam[]) {
    const response = await fetch(`/send-message`, {
      method: "POST",
      body: JSON.stringify({ messages }),
    });

    console.log({ response });
    return response.json();
  }

  async sendMessageToStream(messages: Message[], threadId: string) {
    console.log("sending message to stream");

    // Check if stream is connected first
    const streamService = this.streamService || getStreamService();

    if (streamService) {
      // Check connection and reconnect if needed
      streamService.reconnectIfNeeded();

      // Only proceed if we have a valid connection
      if (!streamService.isStreamConnected()) {
        console.warn(
          "Stream is not connected. Attempting to reconnect before sending.",
        );
        // Give a short time for reconnection
        await new Promise((resolve) => setTimeout(resolve, 500));

        // Check again
        if (!streamService.isStreamConnected()) {
          console.error("Failed to reconnect stream. Cannot send message.");
          return false;
        }
      }
    }

    try {
      // Create a FormData object
      const formData = new FormData();

      // Convert messages to a format without blob data
      const messagesForJson = messages.map((message) => {
        const messageCopy = {
          ...message,
          attachments:
            message.attachments?.map((attachment) => ({
              id: attachment.id,
              type: attachment.type,
              name: attachment.name || "",
              // Binary data will be sent separately
            })) || [],
        };

        return messageCopy;
      });

      // Add messages data as JSON
      formData.append(
        "messagesData",
        JSON.stringify({
          messages: messagesForJson,
          threadId,
        }),
      );

      // Add each message's attachments to the FormData
      messages.forEach((message) => {
        if (message.attachments?.length) {
          message.attachments.forEach((attachment) => {
            // Create a unique key for each attachment using message ID and attachment ID
            const attachmentKey = `attachment_${message.id}_${attachment.id}`;
            formData.append(
              attachmentKey,
              attachment.binary,
              attachment.name || `file_${attachment.id}`,
            );
          });
        }
      });

      const response = await fetch(
        `/chat-stream?thread_id=${encodeURIComponent(threadId)}`,
        {
          method: "POST",
          // Don't set Content-Type header - the browser will set it with the boundary
          body: formData,
        },
      );

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(
          `Failed to send message: ${response.status} ${errorText}`,
        );
      }

      return true;
    } catch (error) {
      console.error("Error sending message:", error);
      return false;
    }
  }

  async handleUserSend(params: {
    threadId?: string;
    content: string;
    role: MessageRole;
    modelParams?: ModelParams;
    attachments?: Attachment[];
  }) {
    // If no threadId is provided, create a new thread
    // const isNewThread = !params.threadId;
    const threadId = params.threadId || (await this.createThread());

    const name = this.generateThreadName(threadId, params.content);
    console.log({ name });

    // if(isNewThread) {
    //   const name = await this.generateThreadName(threadId, params.content);
    //   this.renameThread(threadId, name);
    // }
    const thread = await this.repository.getThreadById(threadId);
    if (!thread) throw new Error("Thread not found");

    thread.updateLastMessageTime();
    this.repository.saveThread(thread);

    const messages = await this.repository.getMessagesByThreadId(threadId);
    const draftMessage = messages.find(
      (m) => m.status === messageStatusMap.DRAFT,
    );

    // Create and save message
    const message = new Message({
      id: generateUUID(),
      ...draftMessage,
      ...params,
      threadId: threadId,
      status: messageStatusMap.DONE,
      createdAt: new Date(),
    });

    await this.repository.saveMessage(message);
    await this.createAssistantStreamMessageHolder(threadId);

    const messagesToSend =
      await this.repository.getMessagesByThreadId(threadId);
    console.log("Messages to send:", messagesToSend);
    console.log(
      "temperature:",
      messages.map((m) => m.modelParams),
    );

    // const success = await this.sendMessageToStream(messagesToSend, threadId);
    const success = await this.sendMessageToStream(messagesToSend, threadId);
    if (!success) {
      // Handle failure - update UI to show error
      const assistantMessages =
        await this.repository.getMessagesByThreadId(threadId);
      const streamingMessage = assistantMessages.find(
        (m) =>
          m.role === "assistant" && m.status === messageStatusMap.STREAMING,
      );

      if (streamingMessage) {
        streamingMessage.content =
          "Sorry, I encountered an error processing your request.";
        streamingMessage.status = messageStatusMap.ERROR;
        await this.saveMessage(streamingMessage);
      }
    }

    console.log("Message send process completed");
  }

  async getDraftMessage(threadId: string): Promise<Message | undefined> {
    const messages = await this.repository.getMessagesByThreadId(threadId);
    return messages.find((m) => m.status === messageStatusMap.DRAFT);
  }

  saveMessage(message: Message) {
    return this.repository.saveMessage(message);
  }

  createDraftMessage(threadId: string) {
    if (!threadId) throw new Error("Thread ID is required");
    return new Message({
      id: generateUUID(),
      content: "",
      threadId: threadId,
      role: "user",
      status: messageStatusMap.DRAFT,
      createdAt: new Date(),
    });
  }

  async createAssistantStreamMessageHolder(threadId: string) {
    if (!threadId) throw new Error("Thread ID is required");
    const msg = new Message({
      id: generateUUID(),
      content: "",
      threadId: threadId,
      role: "assistant",
      status: messageStatusMap.STREAMING,
      createdAt: new Date(),
    });

    return this.saveMessage(msg);
  }

  subscribeToMessages(
    threadId: string,
    callback: (messages: Message[]) => void,
  ) {
    return this.repository.subscribeToMessages(threadId, callback);
  }

  subscribeToThreads(callback: (threads: Thread[]) => void) {
    return this.repository.subscribeToThreads(callback);
  }
}
