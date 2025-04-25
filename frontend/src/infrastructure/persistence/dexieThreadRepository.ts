import { Message, Thread, ThreadRepository } from "../../domain";
import { DbMessage, DbThread } from "./db-models";
import { db } from "./db-setup";
import { liveQuery } from "dexie";

const sortByCreatedAt = (a: DbMessage, b: DbMessage) =>
  a.createdAt < b.createdAt ? -1 : 1;

export class DexieThreadRepository implements ThreadRepository {
  async getMessagesByThreadId(threadId: string): Promise<Message[]> {
    const dbMessages = await db.messages
      .where("threadId")
      .equals(threadId)
      .toArray();

    return dbMessages.map(this.mapDbMessageToMessage).sort(sortByCreatedAt);
  }

  async saveMessage(message: Message): Promise<void> {
    const dbMessage: DbMessage = message;

    await db.messages.put(dbMessage);
  }

  async saveThread(thread: Thread): Promise<void> {
    const dbThread: DbThread = {
      id: thread.id,
      title: thread.title,
      lastMessageAt: thread.lastMessageAt,
      createdAt: thread.createdAt,
    };

    db.threads.put(dbThread);
  }

  async getThreadById(id: string): Promise<Thread | undefined> {
    const dbThread = await db.threads.get(id);
    if (!dbThread) return undefined;
    return this.mapToThread(dbThread);
  }

  async listThreads(): Promise<Thread[]> {
    const dbThreads = await db.threads.toArray();
    return dbThreads.map(this.mapToThread);
  }

  // Combined operations
  async deleteThreadWithMessages(threadId: string): Promise<void> {
    await db.transaction("rw", db.messages, db.threads, async () => {
      await db.messages.where("threadId").equals(threadId).delete();
      await db.threads.delete(threadId);
    });
  }

  subscribeToThreads(callback: (threads: Thread[]) => void): () => void {
    const threadsObservable = liveQuery(() => db.threads.toArray());

    const subscription = threadsObservable.subscribe({
      next: (dbThreads) => {
        const threads = dbThreads.map(this.mapToThread);
        callback(threads);
      },
      error: (error) => {
        console.error("Error fetching threads:", error);
      },
    });

    return () => {
      subscription.unsubscribe();
    };
  }

  subscribeToMessages(
    threadId: string,
    callback: (messages: Message[]) => void,
  ): () => void {
    const messagesObservable = liveQuery(() =>
      db.messages.where("threadId").equals(threadId).toArray(),
    );

    const subscription = messagesObservable.subscribe({
      next: (dbMessages) => {
        const messages = dbMessages
          .sort(sortByCreatedAt)
          .map(this.mapDbMessageToMessage);
        callback(messages);
      },
      error: (error) => {
        console.error("Error fetching messages:", error);
      },
    });

    return () => {
      subscription.unsubscribe();
    };
  }

  private mapToThread(dbThread: DbThread): Thread {
    return new Thread({ ...dbThread });
  }

  private mapDbMessageToMessage(dbMsg: DbMessage): Message {
    return new Message({
      id: dbMsg.id,
      content: dbMsg.content,
      role: dbMsg.role,
      threadId: dbMsg.threadId,
      createdAt: dbMsg.createdAt,
      status: dbMsg.status,
      attachments: dbMsg.attachments,
    });
  }
}
