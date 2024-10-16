import db, {Message, SavedConversation} from "./db";
import {ChatCompletionMessageParam} from "openai/src/resources/chat/completions";
import axios from "axios";

class Conversation {
    id
    data
    messages

    constructor(c: SavedConversation, messages: Message[]) {
        this.id = c.id
        this.data = c
        this.messages = messages
    }

    async getAllConversations() {
        return db.conversation.list()
    }

    getMessages() {
        return db.messages.getByConversation(this.id);
    }

    async drawMessages() {
       const messages = await db.messages.getByConversation(this.id);
       const rootEl = document.querySelector('#messageRoot') || document.querySelector('#inner');
       console.log({messages, rootEl})
       if (!rootEl) {
           console.log('no root found for messages')
           return;
       }

       rootEl.innerHTML = ''

        console.log({messages})
        // const messages = ['first', 'second', 'third']
       messages.map(m => this.addMessageToDOM(m))
    }


    async addMessageToDOM(message: Message) {
        const rootEl = document.querySelector('#messageRoot');
        if (!rootEl) return;
        // const el =  document.createElement('user-message')
        const div = document.createElement("div");
        if (message.role === 'user') div.innerHTML = `<user-message>${message.content}</user-message>`;
        if (message.role === 'assistant') div.innerHTML = `<assistant-message>${message.content}</assistant-message>`;

        // el.innerText = m
        rootEl?.appendChild(div)
    }

    async handleUserInput(form: HTMLFormElement) {
        console.log("handling")
        //  Get value
        const input = form.querySelector<HTMLInputElement>('#user-input');
        if(!input) return
        const content = input.value;
        if (content.trim() === '') return; // Don't send empty content

        // Save message to db
        const savedUserMessage = await db.messages.create(this.id, {role: 'user', content})
        await this.addMessageToDOM(savedUserMessage)

        // Get past messages
        const messages = await (window as any).goChat.conversation.getMessages()
        const openMessages = messages.map((m: Message) => ({role: m.role, content: m.content}))

        const res = await axios.post(`/send-message`, {messages: openMessages})
        if(res.status === 200) {
            const savedAssistantMessage = await db.messages.create(this.id, {role: 'assistant', content: res.data.content})
            await this.addMessageToDOM(savedAssistantMessage)
        }

        // Clear the input field
        input.value = '';

        // If message is first, make it the title
        if(messages.length === 1) db.conversation.update({...this.data, title: content})
    }

    async handleServerResponse(r: any) {
        const {content} = JSON.parse(r)
        const savedMessage = await db.messages.create(this.id, {role: 'assistant', content})
        await this.addMessageToDOM(savedMessage)
        //  Todo: handle exceptions
    }

    async delete() {
        console.log({deleting: this.id})
        await db.conversation.delete(this.id)
    }

}

export async function initConversation(id: string, isNew: boolean) {
    try {
    const savedConversation = isNew ?  await db.conversation.create(id) : await db.conversation.get(id)
    window.history.replaceState(null, '',window.location.origin + `/c/${id}`);


   //  Todo: redirect to 404 page or something else
   if(!savedConversation) {
       return console.log("NO CONVERSATION FOUND FOR:", id)
   }

    if(!savedConversation) throw Error("failed saving conversation")
    const messages = await db.messages.getByConversation(id);
    console.log({1: messages})
    const conversation = new Conversation(savedConversation, messages);
    (window as any).goChat.conversation = conversation

   await conversation.drawMessages()
    } catch (err) {
        console.log({err})
    }

}