import db, {getConversation, Message, SavedConversation} from "./db";
import axios from "axios";
import { nanoid } from "nanoid";
import { marked } from "marked";
import {ChatCompletionMessageParam, ChatCompletionUserMessageParam} from "openai/src/resources/chat/completions";

class Conversation {
    id
    data
    messages
    files: string[]

    constructor(c: SavedConversation, messages: Message[]) {
        this.id = c.id
        this.data = c
        this.messages = messages
        this.files = []
    }

    async getAllConversations() {
        return db.conversation.list()
    }

    async addFile(file: File) {
        const data = await axios.post(`/file/upload`, file, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        })

        console.log({data})
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

   async  drawMessages() {
       const messages = await db.messages.getByConversation(this.id);
       const rootEl = document.querySelector('#messageRoot') || document.querySelector('#inner');
       // console.log({messages, rootEl})

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
        const openMessages = messages.map((m: Message) => ({role: m.role, content: m.content})) as ChatCompletionMessageParam[]
        const contextMessages = openMessages.slice(-5)

        // get files
        const conv = await db.conversation.get(this.id)
        const hasFiles = !!conv?.files.map(f => f.id).length

        const answer = await getCompletion(contextMessages, this.id, hasFiles)

        // Hacky, the addMessagToDOM wants a database message but why should we wait until it's saved in the DB?
        this.addMessageToDOM({role: 'assistant', content: answer, conversationId: this.id, id: '0', timestamp: 0})
        await db.messages.create(this.id, {role: 'assistant', content: answer}),
        this.setLoading(false)
    }

}

export async function uploadFile(file: File, conversationId: string) {
    if(!file) {
        throw new Error("missing param file")
    }
    if(!conversationId) {
        throw new Error("missing param conversationId")
    }
    const { data } = await axios.post(`/file/upload`, { file, conversationId }, {
        headers: {
            'Content-Type': 'multipart/form-data'
        }
    })
    const fileId = data.id

    // Todo, maybe we're doing too much here
    const conv = await db.conversation.get(conversationId)
    if(!conv) return null
    await db.conversation.update({...conv, files: [...conv.files, {id: fileId, name: file.name}]})
    return fileId
}

export async function removeFile(fileId: string, conversationId: string) {
    const { data } = await axios.post(`/file/delete`, { fileId, conversationId }, {
        headers: {
            'Content-Type': 'multipart/form-data'
        }
    })

    const conv = await db.conversation.get(conversationId)
    if(!conv) return;
    await db.conversation.update({...conv, files: [...conv.files.filter(f => f.id !== fileId)]})
}

export async function removeConversation(conversationId: string) {
    axios.post(`/conversation/delete`, { conversationId }, {
        headers: {
            'Content-Type': 'multipart/form-data'
        }
    })
    db.conversation.delete(conversationId);
}

export async function createInStorage() {
    const id  = nanoid().replace(/-/gi, '');
    db.conversation.create(id)
    return id;
}

export async function createUserMessage(content: string, id: string) {
     db.messages.create(id, {role: 'user', content})
}

async function getCompletion(messages: ChatCompletionMessageParam[], conversationId?: string, hasFiles?: boolean): Promise<string> {
    try {
        const res = await axios.post(`/send-message`, {messages, conversationId, hasFiles})
        return res.data.content
    } catch (err) {
        console.log(err)
        return "Oeps, er is iets mis. We sturen er een ontwikkelaar op af"
    }
}


export async function initConversation(id: string) {
    const savedConversation = await db.conversation.get(id)
    window.history.replaceState(null, '',window.location.origin + `/c/${id}`);


   //  Todo: redirect to 404 page or something else
   if(!savedConversation) {
       return console.log("NO CONVERSATION FOUND FOR:", id)
   }

    if(!savedConversation) throw Error("failed saving conversation")
    const messages = await db.messages.getByConversation(id);


    const lastMessage = messages.slice(-1, messages.length)[0]
    const conversation = new Conversation(savedConversation, messages);
    await conversation.drawMessages()


    if(lastMessage?.role === "user") {
        try {
            conversation.setLoading(true)
            const conv = await db.conversation.get(id)
            const hasFiles = !!conv?.files.map(f => f.id).length
            const openMessages = messages.map(m => ({role: m.role, content: m.content} as ChatCompletionUserMessageParam))
            const answer = await getCompletion(openMessages, id, hasFiles)
            const savedAssistantMessage = await db.messages.create(conversation.id, {role: 'assistant', content: answer})
            conversation.setLoading(false)
            await conversation.addMessageToDOM(savedAssistantMessage)

        } catch (err) {
            console.log({err})
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
        const title= await getCompletion([{role: 'user', content: prompt}])
        db.conversation.update({...savedConversation, title})
    }
    (window as any).goChat.conversation = conversation

   await conversation.drawMessages()


}

