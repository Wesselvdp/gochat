import { Attachment } from "../../domain";

export interface ModelParams {
  temperature?: number;
  top_p?: number;
}

export interface DbMessage {
  id: string;
  content: string;
  status: "draft" | "done" | "streaming" | "error";
  role: "user" | "assistant" | "system";
  threadId: string;
  createdAt: Date;
  attachments?: Attachment[];
  modelParams?: ModelParams;
}

export interface DbThread {
  id: string;
  title: string;
  lastMessageAt?: Date;
  createdAt: Date;
}
