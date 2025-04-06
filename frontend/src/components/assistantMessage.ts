import { LitElement, css, html, unsafeCSS } from "lit";

import { Marked } from "marked";
import { markedHighlight } from "marked-highlight";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { customElement, property, state } from "lit/decorators.js";
import globalStyles from "../styles.scss?inline";
import htmx from "htmx.org";
import hljs from "highlight.js";

const marked = new Marked(
  markedHighlight({
    emptyLangClass: "hljs",
    langPrefix: "hljs language-",
    highlight(code, lang, info) {
      const language = hljs.getLanguage(lang) ? lang : "plaintext";
      return hljs.highlight(code, { language }).value;
    },
  }),
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

// Add this interface to define public methods
export interface AssistantMessageElement extends HTMLElement {
  markdownContent: string;
  parsedMarkdown: string;
  streaming: boolean;
  id: string;

  appendContent(chunk: string): void;

  isDone(id: string): void;
}

@customElement("assistant-message")
export class AssistantMessage
  extends LitElement
  implements AssistantMessageElement
{
  static styles = [
    unsafeCSS(globalStyles),
    css`
      .fade-in {
        animation: fadeIn 0.3s ease-in;
      }
      @keyframes fadeIn {
        from {
          opacity: 0.7;
        }
        to {
          opacity: 1;
        }
      }

      .loading-spinner {
        display: inline-block;
        width: 20px;
        height: 20px;
        border: 2px solid rgba(0, 0, 0, 0.1);
        border-radius: 50%;
        border-top-color: #767676;
        animation: spin 1s ease-in-out infinite;
      }

      @keyframes spin {
        to {
          transform: rotate(360deg);
        }
      }
    `,
  ];

  @property({ type: String })
  markdownContent = "";

  @property({ type: String })
  parsedMarkdown = "";

  @property({ type: String, reflect: true })
  id = "";

  @property({ type: Boolean })
  streaming = false;

  async firstUpdated() {
    // If you want to initially populate from innerHTML
    if (this.innerHTML.trim()) {
      this.markdownContent = this.innerHTML;
      this.innerHTML = "";

      // If this is a loading message, don't parse it as markdown
      if (this.markdownContent === "loading") {
        this.parsedMarkdown = '<div class="loading-spinner"></div>';
      } else {
        await this.parseMarkdown();
      }
    }
  }

  async updated(changedProperties: any) {
    if (changedProperties.has("markdownContent") && this.markdownContent) {
      await this.parseMarkdown();
    }

    if (!this.shadowRoot) return;

    const htmxElement = this.shadowRoot.querySelector(
      '[hx-get="/component/avatar"]',
    );
    if (!htmxElement) return;

    // Trigger an HTTP request directly
    fetch("/component/avatar")
      .then((response) => response.text())
      .then((htmlContent) => {
        // Update the element's content
        htmxElement.innerHTML = htmlContent;
      });
  }

  async parseMarkdown() {
    this.parsedMarkdown = await marked.parse(this.markdownContent, {
      breaks: true,
      gfm: true,
    });
  }

  async isDone(id: string) {
    this.id = id;
    this.streaming = false;
  }

  // Add a public method to append content for streaming
  appendContent(chunk: string) {
    // If this is the first content and we're replacing "loading"
    if (
      this.markdownContent === "loading" ||
      this.markdownContent === ". . ."
    ) {
      this.markdownContent = chunk;
    } else {
      this.markdownContent += chunk;
    }

    this.streaming = true;
    // Request an update to trigger the updated lifecycle method
    this.requestUpdate();
  }

  render() {
    return html`
      <div class="flex px-5 gap-6 max-w-5xl mb-2">
        <div>
          <div hx-get="/component/avatar"></div>
        </div>
        <div>
          <p class="font-bold mb-4 text-heading">AĿbert</p>
          <div
            class="max-w-2xl markdown-content ${this.streaming
              ? "fade-in"
              : ""}"
          >
            ${unsafeHTML(this.parsedMarkdown)}
          </div>
        </div>
      </div>
    `;
  }
}
