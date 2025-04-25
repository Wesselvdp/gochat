import { LitElement, css, html, unsafeCSS } from "lit";
import { customElement, property } from "lit/decorators.js";
import globalStyles from "../styles.scss?inline";

// import { removeConversation } from "../conversation";

@customElement("recent-conversation")
export class RecentConversation extends LitElement {
  static styles = [unsafeCSS(globalStyles)];

  @property() id = "";
  @property() title = "";

  constructor() {
    super();
    this.id = this.getAttribute("id") || "";
    this.title = this.getAttribute("title") || "";
  }

  updated() {
    htmx.process(this.shadowRoot);
  }
  render() {
    return html`
      <div
        hx-get="/thread/${this.id}"
        hx-target="global #inner"
        hx-push-url="true"
        class="group p-2 cursor-pointer rounded-lg hover:bg-slate-900/50 opacity-70 hover:opacity-100 transition-all flex gap-2 items-center max-w-2xl"
      >
        <div class="flex items-center gap-3 max-w-[80%]">
          <!--          <span class="material-symbols-outlined text-xs">forum </span>-->
          <span
            class="overflow-ellipsis overflow-hidden whitespace-nowrap text-sm"
            >${this.title}</span
          >
        </div>

        <span
          id="deleteBtn"
          @click="${this.handleDelete}"
          class="group-hover:opacity-100 opacity-0 material-symbols-outlined text-sm ml-auto"
        >
          delete
        </span>
      </div>
    `;
  }

  private handleDelete(e: Event) {
    e.stopPropagation();

    // removeConversation(this.id);

    if (window.location.href.endsWith(`thread/${this.id}`)) {
      // Use HTMX JS API to make request and target element outside shadow DOM
      (window as any).htmx.ajax("GET", "/component/newchat", {
        target: "#inner",
      });
    }
    // Reload recent conversations
    (window as any).goChat.recentConversations.init();
  }
}
