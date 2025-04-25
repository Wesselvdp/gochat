export class Thread {
  id: string;
  title: string;
  lastMessageAt?: Date;
  createdAt: Date;

  constructor(params: {
    id: string;
    title: string;
    lastMessageAt?: Date;
    createdAt: Date;
  }) {
    this.id = params.id;
    this.title = params.title;
    this.lastMessageAt = params.lastMessageAt;
    this.createdAt = params.createdAt;
  }

  rename(newTitle: string) {
    if (!newTitle.trim()) throw new Error("Title cannot be empty");
    this.title = newTitle;
  }

  updateLastMessageTime() {
    this.lastMessageAt = new Date();
  }
}
