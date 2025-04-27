import { LitElement, html, css } from "lit";
import { customElement } from "lit/decorators.js";

@customElement("dialog-title")
export class DialogTitle extends LitElement {
  render() {
    return html`
      <h2
        class="text-lg font-semibold leading-none tracking-tight"
        part="title"
        id="dialog-title"
      >
        <slot></slot>
      </h2>
    `;
  }
}
