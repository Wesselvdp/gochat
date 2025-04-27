export const messageRoleMap = {
  USER: "user",
  ASSISTANT: "assistant",
  SYSTEM: "system",
} as const;

export const messageStatusMap = {
  DRAFT: "draft",
  DONE: "done",
  ERROR: "error",
  STREAMING: "streaming",
} as const;
export type MessageStatus =
  (typeof messageStatusMap)[keyof typeof messageStatusMap];
export type MessageRole = (typeof messageRoleMap)[keyof typeof messageRoleMap];

export type Attachment = {
  id: string;
  name?: string;
  binary: Blob;
  type: string;
};

export type ModelParams = {
  temperature?: number;
  top_p?: number;
};

export class Message {
  id: string;
  role: MessageRole;
  content: string;
  threadId: string;
  createdAt: Date;
  status: MessageStatus;
  attachments: Attachment[];
  modelParams: ModelParams;

  constructor(data: {
    id: string;
    role: MessageRole;
    content: string;
    createdAt: Date;
    threadId: string;
    status: MessageStatus;
    modelParams?: ModelParams;
    attachments?: Attachment[];
  }) {
    this.id = data.id;
    this.role = data.role;
    this.content = data.content;
    this.createdAt = data.createdAt;
    this.threadId = data.threadId;
    this.attachments = data.attachments || [];
    this.status = data.status;
    this.modelParams = data.modelParams || {};
  }

  edit(newContent: string) {
    if (!newContent.trim()) throw new Error("Message content cannot be empty");
    this.content = newContent;
  }

  addAttachment(attachment: Attachment) {
    if (!this.attachments) this.attachments = [];
    this.attachments.push(attachment);
    return this;
  }
}
