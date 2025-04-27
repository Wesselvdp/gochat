import { customElement, state } from "lit/decorators.js";
import { html } from "lit";
import { TailwindElement } from "./albertElement";
import { DexieThreadRepository } from "../infrastructure/persistence/dexieThreadRepository";
import { ChatService } from "../application/ChatService";

import { ModelParams } from "../domain";

@customElement("chat-provider")
export class ChatProvider extends TailwindElement {
  private chatService: ChatService;
  constructor() {
    super();
    const threadRepository = new DexieThreadRepository();
    this.chatService = new ChatService(threadRepository);
  }

  @state()
  modelParams: ModelParams | undefined = {};

  setModelParams(e: CustomEvent) {
    this.modelParams = e.detail;
  }

  sendMessage(e: CustomEvent) {
    return this.chatService.handleUserSend({
      threadId: e.detail.threadId,
      content: e.detail.content,
      role: "user",
      modelParams: this.modelParams,
    });
  }

  render() {
    return html` <slot @set-model-params="${this.setModelParams}"></slot> `;
  }
}
