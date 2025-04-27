import { LitElement, html, css } from "lit";
import { customElement } from "lit/decorators.js";
import { TailwindElement } from "../albertElement";

@customElement("dialog-header")
export class DialogHeader extends TailwindElement {
  // createRenderRoot() { return this; } // Optional: Disable shadow DOM

  render() {
    return html`
      <div
        class="flex flex-col space-y-1.5 text-center sm:text-left"
        part="header"
      >
        <slot></slot>
      </div>
    `;
  }
}
