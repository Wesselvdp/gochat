import {css, html, LitElement, unsafeCSS} from 'lit';


import {customElement, property, state} from 'lit/decorators.js';
import globalStyles from '../styles.scss?inline';

import db from "../db";
import api from "../api";


type FileStatus = 'loading' | 'success' | 'error'
type FileEntry = { name: string, status: FileStatus, id: string }

@customElement('conversation-files')
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
            border: 2px solid #FFF;
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
    ]
    // Declare reactive properties
    @property({type: String})
    conversationId: string = '';


    @property({attribute: false}) // 'attribute: false' prevents Lit from treating it as an attribute    handleMessage: () => null = () => null
    initConversation?: () => Promise<any>

    @state()
    files: FileEntry[] = []

    // Function to fetch items (simulate fetching data)
    async fetchFiles(): Promise<void> {
        const conversation = await db.conversation.get(this.conversationId)
        this.files = conversation?.files.map(f => ({name: f.name, id: f.id, status: 'success'})) || []
    }

    // Call fetchItems() when the component first updates
    firstUpdated() {
        this.fetchFiles();
    }

    private _setLastFile(file: FileEntry) {
        const arr = [...this.files]
        this.files = [...arr.slice(0, -1), file]
    }

    getConversationId() {
        if (this.conversationId) return this.conversationId;
        return this.initConversation?.()
    }

    private async _handleFileSubmit(e: any) {
        const file = e.target.files[0];
        this.files = [...this.files, {name: file.name, status: 'loading', id: ""}]

        const id = await this.getConversationId()

        const conv = await db.conversation.get(id);
        if (!conv) return null;

        try {
            const fileRes = await api.file.create(file, id)
            await db.conversation.update({
                ...conv,
                files: [{id: fileRes.data.id, name: file.name}]
            });
            this._setLastFile({status: 'success', id: fileRes.data.id, name: file.name})
        } catch (err) {
            this.files = [...this.files.slice(0, -1)]
        }

    }

    private async _onFileClick(fileId: string) {
        try {
            await api.file.delete(fileId, this.conversationId)
            this.files = [...this.files.filter(f => f.id !== fileId)]
        } catch (err) {
            console.log({err})
        }
    }


    // Render the UI as a function of component state
    render() {
        return html`
            <div class="bg-background-4 rounded-b-lg px-4 mx-3 mt-[-3px] flex ">
                <div class="flex-1 flex gap-4 py-2">
                    ${this.files.map((file, i) =>
                            html`
                                <div class="group px-4 py-4 my-2 text-sm relative hover:opacity-80 transition-all flex items-center rounded-md bg-background-tertiary cursor-pointer">
                                    ${file.status === 'loading' ? html`<span class="loader"></span>` : ''}
                                    ${file.status === 'success' ? html`<span>${file.name}</span>` : ''}
                                    <div @click="${() => this._onFileClick(file.id)}"
                                         class="group-hover:opacity-100 opacity-0 transition-all absolute top-[-10px] left-[-10px]">
                                        <icon-remove></icon-remove>
                                    </div>
                                    <div>
                            `
                    )}
                </div>
                <div class="py-2 flex items-center">
                    <form>
                        <label class="cursor-pointer p-[2px] rounded block hover:bg-slate-500/90 transition-all">
                            <input @change="${this._handleFileSubmit}" type="file"/>
                            <icon-paperclip/>
                        </label>
                    </form>
                </div>
            </div>


        `;
    }
}

// <div class="bg-slate-200 rounded-lg p-4">
// <form action="">
// <span class="material-symbols-outlined">
//
//
//     </span>
//
//     </form>
//     <p>Hello world</p>
// <div class="bg-slate-200">
//     files: ${this.files.map((file) =>
//                             html`<div class="px-4 py-1 bg-slate-400 mx-2">${file}</div> `
// )}
// </div>
// </div>

