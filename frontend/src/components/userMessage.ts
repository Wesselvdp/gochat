class UserMessage extends HTMLElement {
    constructor() {
        super()
        const content = this.textContent;
        this.innerHTML = `
        <div class="w-full">
            <div class="flex flex-col justify-between px-5 mb-3 max-w-5xl mx-auto rounded-lg group">
                <div class="flex justify-end pb-1">
                    <div class="rounded-lg bg-background-bubble px-2 py-2">
                        <p style="">${content}</p>
                    </div>
                </div>
            </div>
        </div>
        `;
    }
}

customElements.define('user-message', UserMessage)
