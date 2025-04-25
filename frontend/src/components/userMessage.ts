import { LitElement, css, html, unsafeCSS } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import globalStyles from "../styles.scss?inline";

@customElement("user-message")
export class userMessage extends LitElement {
  static styles = [unsafeCSS(globalStyles), css``];

  @property() content = "";

  render() {
    return html`<div class="w-full">
      <div
        class="flex flex-col justify-between px-5 mb-3 max-w-5xl mx-auto rounded-lg group"
      >
        <div class="flex justify-end pb-1">
          <div class="rounded-lg bg-background-bubble px-2 py-2">
            <p>${this.content}</p>
          </div>
        </div>
      </div>
    </div>`;
  }
}
