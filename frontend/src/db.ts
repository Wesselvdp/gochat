import { DBSchema, openDB } from "idb";
import { nanoid } from "nanoid";

import {
  ChatCompletionMessageParam,
  ChatCompletionUserMessageParam,
} from "openai/src/resources/chat/completions";

export type Message = ChatCompletionMessageParam & {
  id: string;
  conversationId: string;
  timestamp: number;
};

export type SupportedFileType = "pdf" | "txt" | "jpg" | "png";

export type FileEntry = {
  type: SupportedFileType;
  id: string;
  name: string;
};

export type SavedConversation = {
  title: string;
  id: string;
  date: Date;
  files: FileEntry[];
};

interface MyDB extends DBSchema {
  conversations: {
    key: string;
    value: SavedConversation;
  };
  messages: {
    key: string;
    value: Message;
    indexes: { conversationId: string };
  };
}

// ChatCompletionUserMessageParam
async function init() {
  return openDB<MyDB>("goChat", 1, {
    upgrade(db) {
      // Create a store of objects
      db.createObjectStore("conversations", {
        keyPath: "id",
        autoIncrement: true,
      });
      const messageStore = db.createObjectStore("messages", {
        keyPath: "id",
        autoIncrement: true,
      });

      messageStore.createIndex("conversationId", "conversationId");
    },
  });
}

// Create conversation
export async function createConversation(id: string) {
  const db = await init();
  const value = {
    id,
    date: new Date(),
    title: "Nieuw gesprek",
    files: [],
  };

  await db.add("conversations", value);
  return value;
}

// Get conversation
export async function getConversation(id: string) {
  const db = await init();
  return db.get("conversations", id);
}

// Update conversation
export async function updateConversation(c: SavedConversation) {
  const db = await init();
  return db.put("conversations", c);
}

export async function deleteConversation(id: string) {
  const db = await init();
  return db.delete("conversations", id);
}

// Get conversation
export async function listConversations() {
  const db = await init();
  const all = await db.getAll("conversations");

  return all.sort((a, b) => b.date.getTime() - a.date.getTime());
}

// Add message to conversation
export async function createMessage(
  conversationId: string,
  message: ChatCompletionMessageParam,
) {
  const db = await init();
  const value = {
    id: nanoid(),
    timestamp: Date.now(),
    conversationId,
    ...message,
  };

  await db.add("messages", value);

  return value;
}

// Update an existing message in the conversation
export async function updateMessage(
  messageId: string,
  updates: { content: string },
) {
  const db = await init();

  // Get the existing message first
  const existingMessage = (await db.get("messages", messageId)) as Message;

  if (!existingMessage) {
    throw new Error(`Message with ID ${messageId} not found`);
  }

  // Type guard to ensure we're dealing with a message with a string content
  if (typeof existingMessage.content !== "string") {
    throw new Error("Cannot update a message with non-string content");
  }

  // Create a new message with updated content
  // We need to explicitly specify content as a string to satisfy the Message type
  const updatedMessage: Message = {
    ...existingMessage,
    content: updates.content,
  };

  // Update the message in the database
  await db.put("messages", updatedMessage);

  return updatedMessage;
}

export async function getMessagesForConversation(conversationId: string) {
  const db = await init();

  try {
    // Use the index we created on conversationId
    const messages: Message[] = await db.getAllFromIndex(
      "messages",
      "conversationId",
      conversationId,
    );

    // Sort messages by timestamp if needed
    messages.sort((a, b) => a.timestamp - b.timestamp);

    return messages;
  } catch (error) {
    console.error("Error fetching messages:", error);
    return [];
  } finally {
    db.close();
  }
}

export default {
  messages: {
    create: createMessage,
    getByConversation: getMessagesForConversation,
    update: updateMessage,
  },
  conversation: {
    get: getConversation,
    create: createConversation,
    list: listConversations,
    update: updateConversation,
    delete: deleteConversation,
  },
};
