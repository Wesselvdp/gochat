import db, {Message, SavedConversation} from "./db";
import api from "./api";
import {ChatCompletionMessageParam} from "openai/src/resources/chat/completions";

async function uploadFile(file: File, conversationId: string) {
    if(!file) {
        throw new Error("missing param file")
    }
    if(!conversationId) {
        throw new Error("missing param conversationId")
    }
    const { data } = await api.file.create(file, conversationId)
    const fileId = data.id

    // Todo, maybe we're doing too much here
    const conv = await db.conversation.get(conversationId)
    if(!conv) return null
    await db.conversation.update({...conv, files: [...conv.files, {id: fileId, name: file.name}]})
    return fileId
}

async function removeFile(fileId: string, conversationId: string) {
    api.file.delete(fileId, conversationId)
    const conv = await db.conversation.get(conversationId)
    if(!conv) return;
    await db.conversation.update({...conv, files: [...conv.files.filter(f => f.id !== fileId)]})
}

async function getCompletion(messages: ChatCompletionMessageParam[], conversationId?: string, hasFiles?: boolean): Promise<string> {
    try {
        const res = await api.conversation.sendMessage(messages, conversationId, hasFiles)
        return res.data.content
    } catch (err) {
        console.log(err)
        return "Oeps, er is iets mis. We sturen er een ontwikkelaar op af"
    }
}

export default {
    getCompletion,
    removeFile,
    uploadFile,
}
