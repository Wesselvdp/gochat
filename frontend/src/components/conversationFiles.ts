import { css, html, LitElement, unsafeCSS } from "lit";

import { customElement, property, state } from "lit/decorators.js";
import globalStyles from "../styles.scss?inline";

import db from "../db";
import api from "../api";
import { DexieThreadRepository } from "../infrastructure/persistence/dexieThreadRepository";
import { ChatService } from "../application/ChatService";
import { Message } from "../domain";
import { generateUUID, toBase64 } from "../utils";

type FileStatus = "loading" | "success" | "error";
type FileEntry = { name: string; status: FileStatus; id: string };

@customElement("conversation-files")
export class SimpleGreeting extends LitElement {
  static styles = [
    unsafeCSS(globalStyles),
    css`
      input[type="file"] {
        display: none;
      }

      .custom-file-upload {
        cursor: pointer;
      }

      .loader {
        width: 20px;
        height: 20px;
        border: 2px solid #fff;
        border-bottom-color: transparent;
        border-radius: 50%;
        display: inline-block;
        box-sizing: border-box;
        animation: rotation 1s linear infinite;
      }
      @keyframes rotation {
        0% {
          transform: rotate(0deg);
        }
        100% {
          transform: rotate(360deg);
        }
      }
    `,
  ];
  // Declare reactive properties
  @property()
  threadId: string = "";

  @property({ attribute: false })
  createThread?: () => Promise<any>;

  @state()
  attachments: FileEntry[] = [];

  @state()
  draftMessage: Message | undefined = undefined;

  private chatService: ChatService;

  constructor() {
    super();
    const threadRepository = new DexieThreadRepository();
    this.chatService = new ChatService(threadRepository);
  }

  // Function to fetch items (simulate fetching data)
  async fetchAttachments(): Promise<void> {
    const attachments = await this.chatService.getAttachements(this.threadId);

    this.attachments =
      attachments.map(
        (a) =>
          ({
            name: a.name || "geen_naam",
            id: a.id,
            status: "success",
          }) as FileEntry,
      ) || [];
  }

  // Call fetchItems() when the component first updates
  firstUpdated() {
    this.fetchAttachments();
  }

  private setLastFile(file: FileEntry) {
    const arr = [...this.attachments];
    this.attachments = [...arr.slice(0, -1), file];
  }

  async _getThreadId() {
    if (this.threadId) return this.threadId;
    if (this.createThread) return await this.createThread();
    return ""; // Or throw an error if neither is available
  }

  async _getDraftMessage() {
    if (this.draftMessage) return this.draftMessage;
    const alreadyStoredDraft = await this.chatService.getDraftMessage(
      this.threadId,
    );
    if (alreadyStoredDraft) {
      this.draftMessage = alreadyStoredDraft;
      return alreadyStoredDraft;
    }
    const draftMessage = this.chatService.createDraftMessage(this.threadId);
    this.draftMessage = draftMessage;
    return draftMessage;
  }

  async upload(file: File, threadId: string) {
    return api.file.create(file, threadId);
  }

  handleFileSubmit = async (e: any) => {
    const file = e.target.files[0] as File;
    this.attachments = [
      ...this.attachments,
      { name: file.name, status: "loading", id: "" },
    ];
    const threadId = await this._getThreadId();
    const draftMessage = await this._getDraftMessage();

    try {
      const fileId = generateUUID();
      if (!file.type.startsWith("image/")) {
        await this.upload(file, threadId);
      }

      draftMessage.addAttachment({
        id: fileId,
        name: file.name,
        binary: file,
        type: file.type,
      });

      await this.chatService.saveMessage(draftMessage);

      this.setLastFile({
        status: "success",
        id: fileId,
        name: file.name,
      });
    } catch (err) {
      this.attachments = [...this.attachments.slice(0, -1)];
    }
  };

  private async onFileClick(fileId: string) {
    try {
      await api.file.delete(fileId, this.threadId);
      this.attachments = [...this.attachments.filter((f) => f.id !== fileId)];
    } catch (err) {
      console.log({ err });
    }
  }

  // Render the UI as a function of component state
  render() {
    return html`
      <div class="bg-background-4 rounded-b-lg px-4 mx-3 mt-[-3px] flex ">
        <div class="flex-1 flex gap-4 py-2">
          ${this.attachments.map(
            (file, i) => html`
              <div
                class="group px-4 py-4 my-2 text-sm relative hover:opacity-80 transition-all flex items-center rounded-md bg-background-tertiary cursor-pointer"
              >
                ${file.status === "loading"
                  ? html`<span class="loader"></span>`
                  : ""}
                ${file.status === "success"
                  ? html`<span>${file.name}</span>`
                  : ""}
                <div
                  @click="${() => this.onFileClick(file.id)}"
                  class="group-hover:opacity-100 opacity-0 transition-all absolute top-[-10px] left-[-10px]"
                >
                  <icon-remove></icon-remove>
                </div>
                <div></div>
              </div>
            `,
          )}
        </div>
        <div class="py-2 flex items-center">
          <form>
            <label
              class="cursor-pointer p-[2px] rounded block hover:bg-slate-500/90 transition-all"
            >
              <input @change="${this.handleFileSubmit}" type="file" />
              <icon-paperclip />
            </label>
          </form>
        </div>
      </div>
    `;
  }
}
