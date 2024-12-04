import db, {Message, SavedConversation} from "./db";
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
        // window.scrollTo(0, document.body.scrollHeight);
    }



    async handleUserInput(content: string) {
        if (content.trim() === '') return; // Don't send empty content

        // Save message to db
        const savedUserMessage = await db.messages.create(this.id, {role: 'user', content})
        this.addMessageToDOM(savedUserMessage)
        this.setLoading(true)

        // Get past messages
        const messages = await this.getMessages()
        const openMessages = messages.map((m: Message) => ({role: m.role, content: m.content}))
        const contextMessages = openMessages.slice(-5)

        // get files
        const conv = await db.conversation.get(this.id)
        const hasFiles = !!conv?.files.map(f => f.id).length

        const res = await axios.post(`/send-message`, {messages: contextMessages, conversationId: this.id, hasFiles})
        // if(res.status === 200) {
            const savedAssistantMessage = await db.messages.create(this.id, {role: 'assistant', content: res.data.content})
        this.setLoading(false)

        await this.addMessageToDOM(savedAssistantMessage)
        // }

        // If message is first, make it the title
        // if(messages.length === 1) db.conversation.update({...this.data, title: content})
    }

    async handleServerResponse(r: any) {
        const { content } = JSON.parse(r)
        const savedMessage = await db.messages.create(this.id, {role: 'assistant', content})
        await this.addMessageToDOM(savedMessage)
        //  Todo: handle exceptions
    }

    async delete() {
        console.log({deleting: this.id})
        await db.conversation.delete(this.id)
    }

    uploadFile(file: File) {
        uploadFile(file, this.id)
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
    console.log({conv})
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

export async function createConversation(content: string, id: string) {
    await db.messages.create(id, {role: 'user', content})
    window.location.href = window.location.origin + `/c/${id}`
}

async function getCompletion(messages: ChatCompletionMessageParam[]): Promise<string> {
    try {
        const res = await axios.post(`/send-message`, {messages})
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
            const answer = await getCompletion( messages.map(m => ({role: m.role, content: m.content} as ChatCompletionUserMessageParam)))
            const savedAssistantMessage = await db.messages.create(conversation.id, {role: 'assistant', content: answer})
            await conversation.addMessageToDOM(savedAssistantMessage)
        } catch (err) {
            console.log({err})
            const savedAssistantMessage = await db.messages.create(conversation.id, {role: 'assistant', content: "Oeps, er ging iets mis."})
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

