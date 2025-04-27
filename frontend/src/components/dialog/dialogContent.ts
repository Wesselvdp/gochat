import { LitElement, html, css } from "lit";
import { customElement } from "lit/decorators.js";
import { TailwindElement } from "../albertElement";

// Assuming you have an icon component or SVG string
const closeIconSVG = html`
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    class="h-4 w-4"
  >
    <line x1="18" y1="6" x2="6" y2="18"></line>
    <line x1="6" y1="6" x2="18" y2="18"></line>
  </svg>
`;

@customElement("dialog-content")
export class DialogContent extends TailwindElement {
  render() {
    // The DialogPortal and DialogOverlay are handled slightly differently.
    // Typically, the overlay and content would be sibling elements managed
    // by a parent controller component or rendered conditionally in the main DOM.
    // This example places the content structure itself.
    // You would typically show/hide this component and the overlay together.
    return html`
      <div
        class="fixed bg-level-1 left-[50%] top-[50%] z-50 grid w-full max-w-xl translate-x-[-50%] translate-y-[-50%] gap-4 border p-6 shadow-lg duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%] data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%] sm:rounded-lg"
        part="content"
        role="dialog"
        aria-modal="true"
        @click=${(e: Event) => e.stopPropagation()}
        aria-labelledby="dialog-title"
        aria-describedby="dialog-description"
      >
        <slot></slot>

        <button
          class="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground"
          part="close-button"
          aria-label="Close"
          @click=${this._handleClose}
        >
          ${closeIconSVG}
          <span class="sr-only">Close</span>
        </button>
      </div>
    `;
  }

  // Placeholder for the close logic you'll implement
  private _handleClose() {
    this.dispatchEvent(
      new CustomEvent("close-dialog", { bubbles: true, composed: true }),
    );
    // You'll likely want to manage the 'open' state externally
    // and hide this component and the overlay.
  }
}
