import {LitElement, css, html, unsafeCSS} from 'lit';


import {customElement, property, state} from 'lit/decorators.js';
import globalStyles from '../styles.scss?inline';
import {  createInStorage, initConversation} from "../conversation";
import db from "../db";




@customElement('user-input-form')
export class userInputForm extends LitElement {
    static styles = [
        unsafeCSS(globalStyles),
        css`
        `,
    ]

    @property({ type: String, reflect: true })
    conversationId = '';


    initConversation = async () => {
        const id = await createInStorage()
        this.conversationId = id
        return id
    }

    handleMsg = async (content: string) => {

        // If there is no conversation initted
        if(!(window as any).goChat.conversation) {
            const id = this.conversationId || await this.initConversation()
            db.messages.create(id, {role: 'user', content})
            return  window.location.href = window.location.origin + `/c/${id}`
        }
        // Existing conversation
         (window as any).goChat.conversation.handleUserInput(content)

    }

    render() {
        return html`
            <div class="max-w-3xl mx-auto w-full">
                <text-area .handleMessage="${this.handleMsg}"></text-area>
                <conversation-files .initConversation=${this.initConversation} .conversationId=${this.conversationId}> </conversation-files>
            </div>
        `;
    }
}


