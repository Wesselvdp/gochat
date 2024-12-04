import {LitElement, css, html, unsafeCSS} from 'lit';

import {customElement, property, state} from 'lit/decorators.js';
import globalStyles from '../styles.scss?inline';

@customElement('text-area')
export class textArea extends LitElement {
    static styles = [
        unsafeCSS(globalStyles),
        css`
        `,
    ]
    // Declare reactive properties
    @property({ attribute: false }) // 'attribute: false' prevents Lit from treating it as an attribute    handleMessage: () => null = () => null
    handleMessage?: (msg: string) => void;


    private autoResize(e: InputEvent) {
        const el = e.target as HTMLTextAreaElement; // Assert the type of e.target        if(!el) return;
        el.style.height = 'auto';
        const growHeight = el.scrollHeight
        el.style.height = `${growHeight > 300 ? 300 : growHeight}px`
    }

    // Submit on enter but not shift enter
    private async handleKeyDown(e: KeyboardEvent) {
        const el = e.target as HTMLTextAreaElement; // Assert the type of e.target        if(!el) return;
        if (!el) return;
        if (e.key === 'Enter' && !e.shiftKey) {
            await this.handleMessage?.(el.value);

            // Hack to properly clear the textfield
            setTimeout(() => {
                el.value = '';
                el.focus();
            }, 0);

            return false;


        }
    }

    render() {
        return html`
            <div>
                <form
                >
                    <textarea @keydown="${this.handleKeyDown}" @input="${this.autoResize}" name="message" class="input shadow w-full bg-background-tertiary" id="growingTextarea" placeholder="Vertel..."></textarea>
                </form>
            </div>
        `;
    }
}


