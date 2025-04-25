import { LitElement, css, html, unsafeCSS } from "lit";

import { customElement, property, state } from "lit/decorators.js";
import globalStyles from "../styles.scss?inline";

import { DexieThreadRepository } from "../infrastructure/persistence/dexieThreadRepository";
import { ChatService } from "../application/ChatService";
import { Message } from "../domain";

@customElement("user-input-form")
export class userInputForm extends LitElement {
  static styles = [unsafeCSS(globalStyles), css``];

  @property({ type: String, reflect: true })
  threadId = "";

  private chatService: ChatService;

  @state()
  message: Message | undefined = undefined;

  constructor() {
    super();
    const threadRepository = new DexieThreadRepository();
    this.chatService = new ChatService(threadRepository);
  }

  async connectedCallback() {
    super.connectedCallback();
    if (this.threadId) {
      this.message = await this.chatService.getDraftMessage(this.threadId);
    }
  }

  saveMessage(content: string) {
    return this.chatService.handleUserSend({
      threadId: this.threadId,
      content,
      role: "user",
    });
  }

  getThreadId() {
    return this.threadId || this.chatService.createThread();
  }

  handleMsg = async (content: string) => {
    if (this.threadId) return this.saveMessage(content);
    this.threadId = await this.chatService.createThread();

    await this.saveMessage(content);
    return (window.location.href =
      window.location.origin + `/thread/${this.threadId}`);
  };

  render() {
    return html`
      <div class="max-w-3xl mx-auto w-full">
        <text-area .handleMessage="${this.handleMsg}"></text-area>
        <conversation-files
          .getThreadId=${this.getThreadId()}
          threadId=${this.threadId}
        ></conversation-files>
      </div>
    `;
  }
}
