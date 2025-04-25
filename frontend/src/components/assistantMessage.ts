import { LitElement, css, html, unsafeCSS } from "lit";
import { Marked } from "marked";
import { markedHighlight } from "marked-highlight";
import { unsafeHTML } from "lit/directives/unsafe-html.js";
import { customElement, property, state } from "lit/decorators.js";
import globalStyles from "../styles.scss?inline";
import hljs from "highlight.js";
import { repeat } from "lit/directives/repeat.js";
const marked = new Marked(
  markedHighlight({
    emptyLangClass: "hljs",
    langPrefix: "hljs language-",
    highlight(code, lang) {
      const language = hljs.getLanguage(lang) ? lang : "plaintext";
      return hljs.highlight(code, { language }).value;
    },
  }),
);

export interface AssistantMessageElement extends HTMLElement {
  markdownContent: string;
  parsedMarkdown: string;
  streaming: boolean;
  id: string;

  appendContent(chunk: string): void;
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
        opacity: 0;
        animation: fadeIn 0.4s ease forwards;
        display: inline-block;
      }

      @keyframes fadeIn {
        to {
          opacity: 1;
        }
      }
    `,
  ];

  @property({ type: String }) markdownContent = "";
  @property({ type: String }) parsedMarkdown = "";
  @property({ type: String, reflect: true }) id = "";
  @property({ type: Boolean, reflect: true }) streaming = false;

  @state() private wordCount = 0;

  connectedCallback() {
    super.connectedCallback();
  }

  async firstUpdated() {
    this.loadAvatar();
  }

  async updated(changedProperties: Map<string, any>) {
    if (changedProperties.has("markdownContent")) {
      await this.parseMarkdown();
    }
  }

  @state()
  private spans: any[] = [];

  async loadAvatar() {
    if (!this.shadowRoot) return;
    const avatarEl = this.shadowRoot.querySelector(
      '[hx-get="/component/avatar"]',
    );
    if (!avatarEl) return;

    const response = await fetch("/component/avatar");
    avatarEl.innerHTML = await response.text();
  }

  async parseMarkdown() {
    const rawHtml = await marked.parse(this.markdownContent, {
      breaks: true,
      gfm: true,
    });

    this.parsedMarkdown = rawHtml;
    return rawHtml;
  }

  @state()
  private words: string[] = [];

  async appendContent(chunk: string) {
    this.streaming = true;

    const html = await marked.parse(this.markdownContent, {
      breaks: true,
      gfm: true,
    });
    const newWords2 = this.wrapWordsInSpans(html).map((el) => el.outerHTML);

    this.markdownContent += chunk;
    this.words = newWords2;
  }

  /**
   * Wraps every word within the text nodes of an HTML string in <span> tags with the class 'fade-in'.
   * Preserves the original HTML structure within a temporary container.
   *
   * @param {string} htmlString The HTML string to process.
   * @returns {HTMLSpanElement[]} An array containing all the created <span> elements.
   */
  wrapWordsInSpans(htmlString: string) {
    // Create a temporary div to hold the parsed HTML
    const container = document.createElement("div");
    container.innerHTML = htmlString;

    // Use a TreeWalker to efficiently traverse text nodes
    const walker = document.createTreeWalker(
      container,
      NodeFilter.SHOW_TEXT, // Only show text nodes
      null, // No custom filter needed
    );

    const nodesToProcess = [];
    let node;

    // Collect all text nodes first to avoid issues with modifying the DOM while traversing
    while ((node = walker.nextNode())) {
      nodesToProcess.push(node);
    }

    // Array to store all created span elements
    const allSpans: HTMLSpanElement[] = [];

    nodesToProcess.forEach((textNode) => {
      const parent = textNode.parentNode;
      const text = textNode.nodeValue;

      // Use a regular expression to split words and capture spaces/punctuation
      // This regex splits by whitespace but keeps the whitespace in the results
      const wordsAndSeparators = text?.split(/(\s+)/);

      const fragment = document.createDocumentFragment();

      wordsAndSeparators?.forEach((part) => {
        if (part.match(/\s+/)) {
          // If it's whitespace, just append it as a text node
          fragment.appendChild(document.createTextNode(part));
        } else if (part.length > 0) {
          // If it's a word, wrap it in a span
          const span = document.createElement("span");
          span.textContent = part;
          fragment.appendChild(span);
          // Add the created span to our collection array
          allSpans.push(span);
        }
        // Ignore empty strings that might result from the split
      });

      // Replace the original text node with the new fragment containing spans and text nodes
      // parent.replaceChild(fragment, textNode);
    });

    // Return the array of all created span elements
    // Note: The original HTML structure is modified within the temporary 'container' element,
    // but the function now returns the flat array of spans for easier iteration/animation.
    return allSpans;
  }
  // style="animation-delay: ${delay}s

  render() {
    const staggered = this.words.map((word, i) => {
      return html`<span class="fade-in" "
        >${unsafeHTML(word)}</span
      > `;
    });
    return html`
      <div class="flex px-5 gap-6 max-w-5xl mb-2">
        <div style="width: 34px; height: 34px;">
          <div hx-get="/component/avatar"></div>
        </div>
        <div>
          <p class="font-bold mb-4 text-heading">AĿbert</p>
          <div class="max-w-2xl markdown-content">
            ${this.streaming ? staggered : unsafeHTML(this.parsedMarkdown)}
          </div>
        </div>
      </div>
    `;
  }
}
