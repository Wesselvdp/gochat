import { Attachment } from "../../domain";

export interface DbMessage {
  id: string;
  content: string;
  status: "draft" | "done" | "streaming" | "error";
  role: "user" | "assistant" | "system";
  threadId: string;
  createdAt: Date;
  attachments?: Attachment[];
}

export interface DbThread {
  id: string;
  title: string;
  lastMessageAt?: Date;
  createdAt: Date;
}
