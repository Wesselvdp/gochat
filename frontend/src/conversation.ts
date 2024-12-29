import db, {Message, SavedConversation} from "./db";
import { nanoid } from "nanoid";

import {ChatCompletionMessageParam, ChatCompletionUserMessageParam} from "openai/src/resources/chat/completions";
import api from "./api";
import actions from "./actions";

class Conversation {
    id
    data
    messages
    files: string[]

    constructor(c: SavedConversation) {
        const messages = db.messages.getByConversation(c.id);
        this.id = c.id
        this.data = c
        this.messages = messages
        this.files = []
    }

    getMessages() {
        return db.messages.getByConversation(this.id);
    }

    setLoading(isLoading: boolean) {
        if(isLoading) {
           return  this.addMessageToDOM({role: 'assistant', content: '. . .', id: "loadingMessage", timestamp: 0, conversationId: "1"})
        }
        const loadingEl = document.getElementById("loadingMessage")
        loadingEl?.remove()
    }

   async drawMessages() {
       const messages = await db.messages.getByConversation(this.id);
       const rootEl = document.querySelector('#messageRoot') || document.querySelector('#inner');

       if (!rootEl) {
           console.log('no root found for messages')
           return;
       }

       rootEl.innerHTML = ''
        
       messages.map(m => this.addMessageToDOM(m))
    }


    addMessageToDOM(message: Message) {
        const rootEl = document.querySelector('#messageRoot');
        if (!rootEl) return;
        // const el =  document.createElement('user-message')
        const div = document.createElement("div");
        div.setAttribute("id", message.id)
        if (message.role === 'user') div.innerHTML = `<user-message>${message.content}</user-message>`;
        if (message.role === 'assistant') div.innerHTML = `<assistant-message>${message.content}</assistant-message>`;

        rootEl?.appendChild(div)
        const nestedElement = document.querySelector("#scrollContainer")
        nestedElement?.scrollTo(0, nestedElement.scrollHeight);

    }

    async handleUserInput(content: string) {
        if (content.trim() === '') return; // Don't send empty content

        // Save message to db
        const savedUserMessage = await db.messages.create(this.id, {role: 'user', content})
        this.addMessageToDOM(savedUserMessage)
        this.setLoading(true)

        // Get past messages
        const messages = await this.getMessages()
        const contextMessages = messages.slice(-5)
        const openMessages = contextMessages.map((m: Message) => ({role: m.role, content: m.content})) as ChatCompletionMessageParam[]

        // get files
        const conv = await db.conversation.get(this.id)
        const hasFiles = !!conv?.files.map(f => f.id).length

        const answer = await actions.getCompletion(openMessages, this.id, hasFiles)
        // Hacky, the addMessagToDOM wants a database message but why should we wait until it's saved in the DB?
        this.addMessageToDOM({role: 'assistant', content: answer, conversationId: this.id, id: '0', timestamp: 0})
        await db.messages.create(this.id, {role: 'assistant', content: answer})
        this.setLoading(false)
    }

}

export async function removeConversation(conversationId: string) {
    api.conversation.delete(conversationId)
    db.conversation.delete(conversationId);
}

export async function createInStorage() {
    const id  = nanoid().replace(/-/gi, '');
    db.conversation.create(id)
    return id;
}

export async function initConversation(id: string) {
    const savedConversation = await db.conversation.get(id)
    window.history.replaceState(null, '',window.location.origin + `/c/${id}`);

   //  Todo: redirect to 404 page or something else
   if(!savedConversation) {
       return console.log("NO CONVERSATION FOUND FOR:", id)
   }

    if(!savedConversation) throw Error("failed saving conversation")

    const conversation = new Conversation(savedConversation);
    await conversation.drawMessages()
    const messages = await conversation.getMessages()
    const lastMessage = messages.slice(-1, messages.length)[0]

    if(lastMessage?.role === "user") {
        try {
            conversation.setLoading(true)
            const hasFiles = !!savedConversation?.files.map(f => f.id).length
            const openMessages = messages.map(m => ({role: m.role, content: m.content} as ChatCompletionUserMessageParam))
            const answer = await api.conversation.sendMessage(openMessages, id, hasFiles)
            const savedAssistantMessage = await db.messages.create(conversation.id, {role: 'assistant', content: answer.data.content})
            conversation.setLoading(false)
            await conversation.addMessageToDOM(savedAssistantMessage)

        } catch (err) {
            const savedAssistantMessage = await db.messages.create(conversation.id, {role: 'assistant', content: "Oeps, er ging iets mis."})
            conversation.setLoading(false)
            await conversation.addMessageToDOM(savedAssistantMessage)
        }
    }

    if(messages.length === 1) {
        const content = messages[0].content
        const prompt = `
            I want you to summarize the message below into 1 short sentence so it can serve as the title of the conversation the message is opening.
            
            message:
            ${content}
        `
        const res= await api.conversation.sendMessage([{role: 'user', content: prompt}])
        db.conversation.update({...savedConversation, title: res.data.content})
    }
    (window as any).goChat.conversation = conversation

   await conversation.drawMessages()
}

