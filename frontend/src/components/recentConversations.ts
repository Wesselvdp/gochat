import { Conversation, removeConversation } from "../conversation";
import { customElement, property, state } from "lit/decorators.js";

import { LitElement, html, PropertyValues, unsafeCSS, css } from "lit";
import globalStyles from "../styles.scss?inline";
import { ChatService } from "../application/ChatService";
import { DexieThreadRepository } from "../infrastructure/persistence/dexieThreadRepository";
import { Thread } from "../domain";

type GroupedThread = [boolean, Thread];

@customElement("recent-conversations")
export class RecentConversations extends LitElement {
  static styles = [
    unsafeCSS(globalStyles),
    unsafeCSS(`
      @import url('https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200');
    `),
  ];

  private chatService: ChatService;

  constructor() {
    super();
    const threadRepository = new DexieThreadRepository();
    this.chatService = new ChatService(threadRepository);
  }

  @state()
  threads: GroupedThread[] = [];
  @state()
  filter = "";

  isToday(date: Date) {
    const today = new Date();
    return date.setHours(0, 0, 0, 0) === today.setHours(0, 0, 0, 0);
  }

  showThread(conversation: Thread) {
    if (!this.filter) return true;
    return conversation.title.toLowerCase().includes(this.filter);
  }

  groupThread = (thread: Thread): GroupedThread => [
    !!thread.lastMessageAt && this.isToday(thread.lastMessageAt),
    thread,
  ];

  filterBy(query: string) {}

  connectedCallback() {
    super.connectedCallback();
    console.log("start");
    this.chatService.subscribeToThreads((threads) => {
      console.log({ threads });
      this.threads = threads.map(this.groupThread);
    });
  }

  heading(text: string) {
    return html`
      <div class=" mb-2 text-heading px-2 text-xs font-extralight">${text}</div>
    `;
  }

  // Inject the HTML content into the component
  render() {
    const today = this.threads.flatMap(([isToday, c]) =>
      isToday && this.showThread(c)
        ? html` <recent-conversation
            id="${c.id}"
            title="${c.title}"
          ></recent-conversation>`
        : [],
    );

    const recent = this.threads.flatMap(([isToday, c]) =>
      !isToday && this.showThread(c)
        ? html` <recent-conversation
            id="${c.id}"
            title="${c.title}"
          ></recent-conversation>`
        : [],
    );

    return html`
      <div class="flex-grow overflow-y-auto">
        <!--        Search box-->
        <div class="px-2">
          <label>
            <div
              class="flex group focus-within:opacity-80 transition-all opacity-30 items-center justify-center mb-4 border-b border-slate-300"
            >
              <span class="material-symbols-outlined text-sm">search</span>
              <input
                @input="${this.handleSearch}"
                type="text"
                class="flex border-none h-9 w-full bg-transparent px-3 py-1 text-base shadow-sm transition-colors transition-opacity focus:opacity-100 file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-0 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 md:text-sm"
                id="conversationSearch"
              />
            </div>
          </label>
        </div>

        <div class="max-h-[82vh] overflow-scroll relative">
          ${today.length
            ? html` <div class="mb-4">
                ${this.heading("Vandaag")}
                <div class="space-y-2">${today}</div>
              </div>`
            : null}
          ${recent.length
            ? html`
                <div class="mb-4">
                  ${this.heading("Vorige gesprekken")}
                  <div class="space-y-2">${recent}</div>
                </div>
              `
            : null}
        </div>
      </div>
    `;
  }

  private handleSearch(e: Event) {
    // More specifically, you can use
    const inputEvent = e as InputEvent;

    // Or if you need the target value
    const value = (e.target as HTMLInputElement).value;
    this.filter = value;
  }
}
