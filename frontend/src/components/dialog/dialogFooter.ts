import { html } from "lit";
import { customElement } from "lit/decorators.js";
import { TailwindElement } from "../albertElement";

@customElement("dialog-footer")
export class DialogFooter extends TailwindElement {
  render() {
    // Assigning the id here for aria-describedby in my-dialog-content
    return html`
      <div
        class="flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2"
      >
        <slot></slot>
      </div>
    `;
  }
}
