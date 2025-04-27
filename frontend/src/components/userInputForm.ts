import { LitElement, css, html, unsafeCSS } from "lit";

import { customElement, property, state } from "lit/decorators.js";
import globalStyles from "../styles.scss?inline";

import { DexieThreadRepository } from "../infrastructure/persistence/dexieThreadRepository";
import { ChatService } from "../application/ChatService";
import { Message, ModelParams } from "../domain";
import { request } from "axios";

@customElement("user-input-form")
export class userInputForm extends LitElement {
  static styles = [unsafeCSS(globalStyles), css``];

  @property({ type: String, reflect: true })
  threadId = "";

  private chatService: ChatService;
  constructor() {
    super();
    const threadRepository = new DexieThreadRepository();
    this.chatService = new ChatService(threadRepository);
  }
  @state()
  message: Message | undefined = undefined;

  @state()
  modelParams: ModelParams | undefined = undefined;

  async connectedCallback() {
    super.connectedCallback();
    if (this.threadId) {
      this.message = await this.chatService.getDraftMessage(this.threadId);
    }

    // Fetch model params if they're undefined
    if (this.modelParams === undefined) {
      this.modelParams = (await this.chatService.getModelParams(
        this.threadId,
      )) || { temperature: 0.3, top_p: 0.8 };
    }
  }

  saveMessage(content: string) {
    return this.chatService.handleUserSend({
      threadId: this.threadId,
      content,
      role: "user",
      modelParams: this.modelParams,
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

  setModelParams = (modelParams: ModelParams) => {
    console.log("setttingngng!", modelParams);
    this.modelParams = modelParams;
    this.requestUpdate();
  };

  render() {
    return html`
      <div class="max-w-3xl mx-auto w-full">
        <text-area .handleMessage="${this.handleMsg}"></text-area>
        <conversation-files
          .setModelParams=${this.setModelParams}
          .modelParams=${this.modelParams}
          threadId=${this.threadId}
        ></conversation-files>
      </div>
    `;
  }
}
