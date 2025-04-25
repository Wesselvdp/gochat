import { Message } from "./Message";
import { Thread } from "./Thread";

export interface ThreadRepository {
  getMessagesByThreadId(threadId: string): Promise<Message[]>;
  saveMessage(message: Message): Promise<void>;
  deleteThreadWithMessages(threadId: string): Promise<void>;
  getThreadById(id: string): Promise<Thread | undefined>;
  listThreads(): Promise<Thread[]>;
  saveThread(thread: Thread): Promise<void>;
  subscribeToMessages(
    threadId: string,
    callback: (messages: Message[]) => void,
  ): () => void;
  subscribeToThreads(callback: (messages: Thread[]) => void): () => void;
}
