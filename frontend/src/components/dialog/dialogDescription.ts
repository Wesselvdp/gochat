import { html } from "lit";
import { customElement } from "lit/decorators.js";
import { TailwindElement } from "../albertElement";

@customElement("dialog-description")
export class DialogDescription extends TailwindElement {
  render() {
    return html`
      <p
        class="text-sm text-muted-foreground"
        part="description"
        id="dialog-description"
      >
        <slot></slot>
      </p>
    `;
  }
}
