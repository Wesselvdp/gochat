import { customElement, property } from "lit/decorators.js";
import { html } from "lit";
import { TailwindElement } from "./albertElement";

@customElement("torgon-input")
export class TorgonInput extends TailwindElement {
  // Input properties
  @property()
  type = "text";

  @property()
  label = "";

  @property()
  placeholder = "";

  @property()
  value = "";

  @property()
  name = "";

  @property()
  id = "";

  @property()
  required = false;

  @property()
  disabled = false;

  @property()
  autocomplete = "";

  @property()
  ariaLabel = "";

  @property()
  ariaDescribedby = "";

  @property()
  errorMessage = "";

  @property()
  helperText = "";

  // Generate a unique ID for accessibility purposes if none provided
  connectedCallback() {
    super.connectedCallback();
    if (!this.id) {
      this.id = `torgon-input-${Math.random().toString(36).substring(2, 9)}`;
    }
  }

  // Handle input changes
  handleInput(e) {
    this.value = e.target.value;
    // Dispatch custom event for parent components
    this.dispatchEvent(
      new CustomEvent("input-change", {
        detail: {
          value: this.value,
          name: this.name,
        },
        bubbles: true,
        composed: true,
      }),
    );
  } // Handle input changes
  handleBlur(e) {
    this.value = e.target.value;
    // Dispatch custom event for parent components
    this.dispatchEvent(
      new CustomEvent("input-blur", {
        detail: {
          value: this.value,
          name: this.name,
        },
        bubbles: true,
        composed: true,
      }),
    );
  }

  render() {
    // Generate IDs for associated elements
    const inputId = this.id;
    const helperId = `${inputId}-helper`;
    const errorId = `${inputId}-error`;

    // Determine which aria-describedby value to use
    const describedBy =
      [
        this.helperText ? helperId : "",
        this.errorMessage ? errorId : "",
        this.ariaDescribedby || "",
      ]
        .filter(Boolean)
        .join(" ") || undefined;

    return html`
      <div class="flex flex-col w-full gap-2">
        ${this.label
          ? html`
              <label
                for="${inputId}"
                class="text-sm font-medium text-foreground"
              >
                ${this.label}${this.required
                  ? html` <span class="text-destructive">*</span>`
                  : ""}
              </label>
            `
          : ""}

        <input
          id="${inputId}"
          type="${this.type}"
          name="${this.name}"
          placeholder="${this.placeholder}"
          .value="${this.value}"
          ?required="${this.required}"
          ?disabled="${this.disabled}"
          autocomplete="${this.autocomplete}"
          aria-label="${this.ariaLabel || this.label || undefined}"
          aria-describedby="${describedBy}"
          aria-invalid="${this.errorMessage ? "true" : undefined}"
          @blur="${this.handleBlur}"
          @input="${this.handleInput}"
          class="flex h-9 w-full rounded-md border ${this.errorMessage
            ? "border-destructive"
            : "border-input"} 
                bg-transparent px-3 py-1 text-base shadow-sm transition-colors 
                file:border-0 file:bg-transparent file:text-sm file:font-medium 
                file:text-foreground placeholder:text-muted-foreground 
                focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring 
                disabled:cursor-not-allowed disabled:opacity-50 md:text-sm"
        />

        ${this.helperText
          ? html`
              <p id="${helperId}" class="text-sm text-muted-foreground">
                ${this.helperText}
              </p>
            `
          : ""}
        ${this.errorMessage
          ? html`
              <p id="${errorId}" class="text-sm text-destructive">
                ${this.errorMessage}
              </p>
            `
          : ""}
      </div>
    `;
  }
}
