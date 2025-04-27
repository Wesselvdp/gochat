import { customElement, property, state } from "lit/decorators.js";
import { html, PropertyValues } from "lit";
import { TailwindElement } from "./albertElement";

@customElement("dialog-provider")
export class DialogProvider extends TailwindElement {
  @property()
  isOpen = false;

  @state()
  template = html``;

  openDialog(e: CustomEvent) {
    this.isOpen = true;
    this.template = e.detail.template;
  }

  render() {
    return html`
      <slot @opens-dialog="${this.openDialog}"></slot>
      ${this.isOpen
        ? html`
            <div
              @click="${() => (this.isOpen = false)}"
              class="absolute inset-0 bg-level-1 opacity-90 z-10 flex items-center justify-center"
            ></div>
            ${this.template}
          `
        : null}
    `;
  }
}
