import db from "./db";
import {marked} from "marked";
import hljs from 'highlight.js'

import {createInStorage, removeConversation, uploadFile} from "./conversation";
class UserMessage extends HTMLElement {
    constructor() {
        super()
        const content = this.textContent;
        this.innerHTML = `
        <div class="w-full">
            <div class="flex flex-col justify-between px-5 mb-3 max-w-5xl mx-auto rounded-lg group">
                <div class="flex justify-end pb-1">
                    <div class="rounded-lg bg-background-bubble px-2 py-2">
                        <p style="white-space: pre-wrap;">${content}</p>
                    </div>
                </div>
            </div>
        </div>
        `;
    }
}


class RecentConversation extends HTMLElement {
    connectedCallback() {
        const id = this.getAttribute('id') || '';
        const title = this.getAttribute('title') || '';

        // Inject the HTML content into the component
        this.innerHTML = `
            <div hx-get="/c/${id}" hx-target="#inner" class="group cursor-pointer opacity-70 hover:opacity-100 transition-all flex gap-2 items-center mb-1 max-w-2xl">
                <div class="flex items-center gap-3 max-w-[80%]"   hx-target="#inner">
                    <span class="material-symbols-outlined text-sm">
                        forum
                    </span>
                    <span class="overflow-ellipsis overflow-hidden whitespace-nowrap">${title} </span>
                </div>
                
                <span id="deleteBtn" class="group-hover:opacity-100 opacity-0 material-symbols-outlined text-sm ml-auto">
                    delete
                </span>
            </div>
            `;

        this.querySelector('#deleteBtn')?.addEventListener('click', () => {
            removeConversation(id);

            if (window.location.href.endsWith(`c/${id}`)) {
                htmx.ajax('GET', '/component/newchat', {target:'#inner'})
            }
            (window as any).goChat.recentConversations.init()

        });


        // Ensure HTMX processes this new HTML after it has been added to the DOM
        (window as any).htmx.process(this);
    }
}



class AssistantMessage extends HTMLElement {
    async connectedCallback() {
        // Create a temporary pre element to preserve formatting
        const pre = document.createElement('pre');
        pre.style.display = 'none';
        pre.textContent = this.innerHTML;
        this.innerHTML = '';
        this.appendChild(pre);

        // Get the preserved content
        const markdownContent = pre.textContent;
        pre.remove();
        const root = this.attachShadow({ mode: 'closed' })
        // Create the base structure
        root.innerHTML = `
            <div class="flex px-5 gap-6 max-w-5xl mb-2">
                <div>   

                   <div hx-get="/component/avatar" />
                </div>
                <div>
                    <p class="font-bold mb-4 text-heading">Kwiz AI</p>
                    <div class="max-w-2xl markdown-content">${marked.parse(markdownContent, {
                            breaks: true,  // Enable line breaks
                            gfm: true
            
                        })}
                    </div>
                </div>
            </div>
        `;

        htmx.process(root) //



        // Process HTMX
    }
}



// MD options
marked.setOptions({


    async: false,
    breaks: false,
    extensions: null,
    gfm: true,
    hooks: null,
    pedantic: false,
    silent: false,
    tokenizer: null,
    walkTokens: null

});

customElements.define('user-message', UserMessage)
customElements.define('assistant-message_old', AssistantMessage)
customElements.define('recent-conversation', RecentConversation)

