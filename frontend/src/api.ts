import axios from "axios";
import {ChatCompletionMessageParam} from "openai/src/resources/chat/completions";

export default {
    conversation: {
        delete: async (conversationId: string) => {
            try   {
                const res = await axios.post(`/conversation/delete`, { conversationId }, {
                    headers: {
                        'Content-Type': 'multipart/form-data'
                    }
                })
                return res
            } catch (err) {
                console.log('error deleting conversation')
            }

        },
        sendMessage: (messages: ChatCompletionMessageParam[], conversationId?: string, hasFiles?: boolean) => axios.post(`/send-message`, {messages, conversationId, hasFiles})
    },
    file: {
        delete: (fileId: string, conversationId: string) => axios.post(`/file/delete`, { fileId, conversationId }, {
                headers: {
                    'Content-Type': 'multipart/form-data'
                }
            }),
        create: (file: File, conversationId: string) => axios.post(`/file/upload`, { file, conversationId }, {
                headers: {
                    'Content-Type': 'multipart/form-data'
                }
            })
    }
}