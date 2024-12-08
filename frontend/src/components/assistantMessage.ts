import {LitElement, css, html, unsafeCSS} from 'lit';

import {Marked} from "marked";
import { markedHighlight } from 'marked-highlight'
import { unsafeHTML } from "lit/directives/unsafe-html.js"
import {customElement, property, state} from 'lit/decorators.js';
import globalStyles from '../styles.scss?inline';
import htmx from 'htmx.org'
import hljs from "highlight.js";

const marked = new Marked(
    markedHighlight({
        emptyLangClass: 'hljs',
        langPrefix: 'hljs language-',
        highlight(code, lang, info) {
            const language = hljs.getLanguage(lang) ? lang : 'plaintext';
            return hljs.highlight(code, { language }).value;
        }
    })
);

// marked.setOptions({
//     async: false,
//     breaks: false,
//     extensions: null,
//     gfm: true,
//     hooks: null,
//     pedantic: false,
//     silent: false,
//     tokenizer: null,
//     walkTokens: null
// });

@customElement('assistant-message')
export class AssistantMessage extends LitElement {
    static styles = [
        unsafeCSS(globalStyles),
        css`
        `,
    ]

    @property({ type: String })
    markdownContent = '';

    @property({ type: String })
    parsedMarkdown = "";

    async firstUpdated() {
        // If you want to initially populate from innerHTML
        this.markdownContent = this.innerHTML;
        this.innerHTML = '';
        // Trigger HTMX after rendering
        await this.parseMarkdown();

    }

    async updated() {
        if(!this.shadowRoot) return;

        const htmxElement = this.shadowRoot.querySelector('[hx-get="/component/avatar"]');
        if(!htmxElement) return;

// Trigger an HTTP request directly
        fetch('/component/avatar')
            .then(response => response.text())
            .then(htmlContent => {
                // Update the element's content
                htmxElement.innerHTML = htmlContent;
            });
        // htmx.process(htmxElement);
    }

    async parseMarkdown() {
        this.parsedMarkdown =  await marked.parse(this.markdownContent, {
            breaks: true,
            gfm: true,
        });
    }

    render() {
        return html`
            <div class="flex px-5 gap-6 max-w-5xl mb-2">
                <div>
                    <div hx-get="/component/avatar"></div>
                </div>
                <div>
                    <p class="font-bold mb-4 text-heading">Kwiz AI</p>
                    <div class="max-w-2xl markdown-content">
                        ${unsafeHTML(this.parsedMarkdown)}
                    </div>
                </div>
            </div>
        `;
    }
}




