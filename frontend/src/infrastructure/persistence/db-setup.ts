import Dexie, { type EntityTable } from "dexie";
import { DbMessage, DbThread } from "./db-models";

class ChatDatabase extends Dexie {
  messages!: EntityTable<DbMessage, "id">;
  threads!: EntityTable<DbThread, "id">;

  constructor() {
    super("ChatDb");

    this.version(1).stores({
      messages: "++id, threadId, content, conversationId, createdAt",
      threads: "++id, title, lastMessageAt, createdAt",
    });
  }
}

const db = new ChatDatabase();

export { db };
