import { LitElement, css, html, unsafeCSS } from "lit";

import { customElement, property, state } from "lit/decorators.js";
import globalStyles from "../styles.scss?inline";
import {
  createUserMessage,
  createInStorage,
  initConversation,
} from "../conversation";

@customElement("user-input-form")
export class userInputForm extends LitElement {
  static styles = [unsafeCSS(globalStyles), css``];

  @property({ type: String, reflect: true })
  conversationId = "";

  @property({ type: Boolean })
  isLocked = false;

  initConversation = async () => {
    const id = await createInStorage();
    this.conversationId = id;
    return id;
  };

  handleMsg = async (msg: string) => {
    // If there is no conversation initted
    this.isLocked = true;
    if (!(window as any).goChat.conversation) {
      const id = this.conversationId || (await this.initConversation());
      await createUserMessage(msg, id);
      return (window.location.href = window.location.origin + `/c/${id}`);
    }
    // Existing conversation
    await (window as any).goChat.conversation.handleUserInput(msg);
    this.isLocked = false;
  };

  render() {
    return html`
      <div
        class="${this.isLocked
          ? "opacity-50"
          : "opacity-100"} max-w-3xl mx-auto w-full"
      >
        <text-area .handleMessage="${this.handleMsg}"></text-area>
        <conversation-files
          .initConversation=${this.initConversation}
          .conversationId=${this.conversationId}
        ></conversation-files>
      </div>
    `;
  }
}
