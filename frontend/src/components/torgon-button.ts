import { LitElement, html } from "lit";
import { customElement, property } from "lit/decorators.js";
import { TailwindElement } from "./albertElement";
// Optional: If class logic gets complex, consider a utility like clsx
// import { clsx } from 'clsx';

// Define the possible variants and sizes as types
export type ButtonVariant =
  | "default"
  | "destructive"
  | "outline"
  | "secondary"
  | "ghost"
  | "link";
export type ButtonSize = "default" | "sm" | "lg" | "icon";

@customElement("torgon-button") // Changed name slightly to avoid collision
export class TorgonButton extends TailwindElement {
  @property({ type: String })
  variant: ButtonVariant = "default";

  @property({ type: String })
  size: ButtonSize = "default";

  @property({ type: Boolean, reflect: true })
  disabled = false;

  @property({ type: String })
  type: "button" | "submit" | "reset" = "button";

  @property({ type: String, attribute: "aria-label" })
  ariaLabel = "";

  @property()
  class = "";

  // --- No static styles block ---

  // Helper function to get variant classes (mimics cva variant logic)
  private _getVariantClasses(): string {
    switch (this.variant) {
      case "destructive":
        return "bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90";
      case "outline":
        return "border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground";
      case "secondary":
        return "bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80";
      case "ghost":
        return "hover:bg-accent hover:text-accent-foreground";
      case "link":
        return "text-primary underline-offset-4 hover:underline";
      case "default":
      default:
        return "bg-primary text-primary-foreground shadow hover:bg-primary/90";
    }
  }

  // Helper function to get size classes (mimics cva size logic)
  private _getSizeClasses(): string {
    switch (this.size) {
      case "sm":
        return "h-8 rounded-md px-3 text-xs";
      case "lg":
        return "h-10 rounded-md px-8";
      case "icon":
        return "h-9 w-9";
      case "default":
      default:
        return "h-9 px-4 py-2";
    }
  }

  render() {
    // Base classes (from cva definition)
    const baseClasses =
      "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0";

    // Get variant and size specific classes
    const variantClasses = this._getVariantClasses();
    const sizeClasses = this._getSizeClasses();

    // Combine all classes
    // Note: Order might matter slightly if there are overrides, but Tailwind is generally designed well for this.
    // Using clsx here could be slightly safer if complex overrides existed:
    // const allClasses = clsx(baseClasses, variantClasses, sizeClasses);
    const allClasses = `${baseClasses} ${variantClasses} ${sizeClasses} ${this.class}`;

    return html`
      <button
        class=${allClasses}
        ?disabled=${this.disabled}
        type=${this.type}
        aria-label=${this.ariaLabel || undefined}
        part="button"
      >
        <slot></slot>
      </button>
    `;
  }

  /**
   * Override createRenderRoot to render in the light DOM (no Shadow DOM)
   * if you want Tailwind styles applied globally without extra config.
   * However, standard practice for reusable web components is Shadow DOM.
   * If using Shadow DOM (default), ensure Tailwind is configured to scan it.
   */
  // protected createRenderRoot() {
  //   return this; // Render to Light DOM
  // }
}

// Optional: Declare the element type for TypeScript
declare global {
  interface HTMLElementTagNameMap {
    "torgon-button": TorgonButton;
  }
}
