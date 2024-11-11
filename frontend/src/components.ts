import db from "./db";
import {marked} from "marked";
import hljs from 'highlight.js'
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

class TextArea extends HTMLElement {
    constructor() {
        super()

        const isNew = this.getAttribute('isNew') || '';


        this.innerHTML = `
            <form
                class="max-w-3xl mx-auto w-full" 
            >
               <textarea name="message" class="input w-full bg-background-tertiary" id="growingTextarea" placeholder="Vertel..."></textarea>
           </form>
        `;

        const textarea: HTMLTextAreaElement | null = this.querySelector('#growingTextarea');
        const form: HTMLFormElement | null = this.querySelector('form');
        if(!textarea) return;

        textarea.addEventListener('input', autoResize);
        textarea.addEventListener('keydown', handleKeyDown);

        function autoResize() {
            if(!textarea) return;
            textarea.style.height = 'auto';
            textarea.style.height = textarea.scrollHeight + 'px';
        }

        function handleKeyDown(e: any) {
            if (e.key === 'Enter' && !e.shiftKey && textarea) {
                e.preventDefault();
                if(isNew) return (window as any).goChat.createConversation(textarea?.value);
                (window as any).goChat.conversation.handleUserInput(textarea?.value)
                textarea.value = ''
            }
        }


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
           db.conversation.delete(id);
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

        // Create the base structure
        this.innerHTML = `
            <div class="flex px-5 gap-6 max-w-5xl mb-2">
                <div>   
<!--                    <div hx-get="/component/avatar" hx-trigger="load"></div>-->
                    <div class="kwizLogo" style="background-image: url('/static/brand.svg')"></div>

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



        // Find the container for markdown content
        const markdownContainer = this.querySelector('.markdown-content');

        // Apply syntax highlighting to code blocks
            markdownContainer?.querySelectorAll('pre code').forEach((block) => {

                hljs.highlightElement(block as any);
            });


        // Process HTMX
        (window as any).htmx.process(this);
    }
}



// MD options
marked.setOptions({
    // highlight: function(code, language) {
    //     if (language && hljs.getLanguage(language)) {
    //         return hljs.highlight(code, { language: language }).value;
    //     }
    //     return hljs.highlightAuto(code).value;
    // },

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
customElements.define('text-area', TextArea)
customElements.define('assistant-message', AssistantMessage)
customElements.define('recent-conversation', RecentConversation)

