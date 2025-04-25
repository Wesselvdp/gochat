import axios from "axios";
import { ChatCompletionMessageParam } from "openai/src/resources/chat/completions";
import { Message } from "./domain";

export default {
  threads: {
    delete: async (conversationId: string) => {
      try {
        const res = await axios.post(
          `/conversation/delete`,
          { conversationId },
          {
            headers: {
              "Content-Type": "multipart/form-data",
            },
          },
        );
        return res;
      } catch (err) {
        console.log("error deleting conversation");
      }
    },
    getCompletion: (messages: Message[], threadId?: string) =>
      axios.post(`/send-message`, { messages, threadId }),
  },
  file: {
    delete: (fileId: string, conversationId: string) =>
      axios.post(
        `/file/delete`,
        { fileId, conversationId },
        {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        },
      ),
    create: (file: File, conversationId: string) =>
      axios.post(
        `/file/upload`,
        { file, conversationId },
        {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        },
      ),
  },
};
