import {LitElement, css, html, unsafeCSS} from 'lit';


import {customElement, property} from 'lit/decorators.js';
import globalStyles from '../styles.scss?inline';
import {createConversation, createInStorage} from "../conversation";



@customElement('user-input-form')
export class userInputForm extends LitElement {
    static styles = [
        unsafeCSS(globalStyles),
        css`
        `,
    ]

    @property({ type: String})
    conversationId = "";

    private handleMsg = async (msg: string) => {
        const currentConversationId = this.conversationId
        // New conversation
        if(!currentConversationId) {
            const id = await createInStorage()
            createConversation(msg, id)
        }
        // Existing conversation
        if(currentConversationId) {
            (window as any).goChat.conversation.handleUserInput(msg)
        }

    }


    // Render the UI as a function of component state
    render() {
        return html`
            <div class="max-w-3xl mx-auto w-full">
                <text-area .handleMessage="${this.handleMsg}"></text-area>
                <conversation-files conversationId="${this.conversationId}"> </conversation-files>
            </div>
        `;
    }
}


