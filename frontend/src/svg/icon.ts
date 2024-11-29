import { LitElement, html, css } from 'lit';
import {customElement, property} from "lit/decorators.js";
import paperclip from './paperclip.svg?raw'

@customElement('svg-icon')
class IconSVG extends LitElement {
    static styles = css`
    svg {
      /* Your global SVG styles here */
    }
  `;

    @property()
    icon: string = "";

    render() {
        // Assuming you've imported your SVGs as modules
        const iconMap: Record<string, string> = {
            paperclip
        };

        const selectedIcon = iconMap[this.icon];
    console.log({selectedIcon})
        return html`
      <svg>${selectedIcon}</svg>
    `;
    }
}

