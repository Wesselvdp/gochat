import {removeConversation} from "../conversation";


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
                htmx.ajax('GET', '/component/newchat', {target: '#inner'})
            }
            (window as any).goChat.recentConversations.init()

        });


        // Ensure HTMX processes this new HTML after it has been added to the DOM
        (window as any).htmx.process(this);
    }
}

customElements.define('recent-conversation', RecentConversation)