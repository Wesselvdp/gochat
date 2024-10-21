import db from "./db";

class UserMessage extends HTMLElement {
    constructor() {
        super()

        this.innerHTML = `
        <div class="w-full">
            <div class="flex flex-col justify-between px-5 mb-3 max-w-5xl mx-auto rounded-lg group">
                <div class="flex justify-end pb-1">
                    <div class="rounded-lg bg-slate-700 px-2 py-2">
                        <p>${this.innerText}</p>
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

        this.innerHTML = `
            <form
            class="max-w-3xl mx-auto w-full"

            >
               <textarea class="input w-full" id="growingTextarea" placeholder="Type your message..."></textarea>
           </form>
        `;

        const textarea: HTMLTextAreaElement | null = this.querySelector('#growingTextarea');
        const form: HTMLFormElement | null = this.querySelector('form');
        if(!textarea) return;

        textarea?.addEventListener('input', autoResize);
        textarea?.addEventListener('keydown', handleKeyDown);

        function autoResize() {
            if(!textarea) return;
            textarea.style.height = 'auto';
            textarea.style.height = textarea.scrollHeight + 'px';
        }

        function handleKeyDown(e: any) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                console.log('submit');
                (window as any).goChat.createConversation(textarea?.value);
            }
        }

        // function submitText() {
        //     const text = textarea.value.trim();
        //     if (text) {
        //         output.textContent = 'Submitted: ' + text;
        //         textarea.value = '';
        //         autoResize.call(textarea);
        //     }
        // }
    }
}

class RecentConversation extends HTMLElement {
    connectedCallback() {
        const id = this.getAttribute('id') || '';
        const title = this.getAttribute('title') || '';

        // Inject the HTML content into the component
        this.innerHTML = `
            <div class="group cursor-pointer opacity-70 hover:opacity-100 transition-all flex gap-2 items-center mb-1 max-w-2xl">
                <div class="flex items-center gap-3"  hx-get="/c/${id}" hx-target="#inner">
                    <span class="material-symbols-outlined text-sm">
                        forum
                    </span>
                    <span>${title} </span>
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
        connectedCallback() {
            // Inject the HTML content into the component
            this.innerHTML = `
                <div class="flex gap-6 max-w-5xl">
                    <div>	
                        <div hx-get="/component/avatar" hx-trigger="load"></div>
                    </div>
                    <div>
                        <p class="font-bold mb-1">modelname</p>
                        <p class="max-w-2xl">${this.innerText}</p>
                    </div>
                </div>
      `;

            // Ensure HTMX processes this new HTML after it has been added to the DOM
            (window as any).htmx.process(this);
        }
    }






customElements.define('user-message', UserMessage)
customElements.define('text-area', TextArea)
customElements.define('assistant-message', AssistantMessage)
customElements.define('recent-conversation', RecentConversation)