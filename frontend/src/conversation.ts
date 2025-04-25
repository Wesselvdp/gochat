// conversation.ts
import db, { Message, SavedConversation } from "./db";
import axios from "axios";
import { nanoid } from "nanoid";
import { ChatCompletionMessageParam } from "openai/src/resources/chat/completions";
import { Stream, newStream } from "./stream";
import { AssistantMessageElement } from "./components/assistantMessage";

class Conversation {
  id: string;
  data: SavedConversation;
  messages: Message[];
  files: string[];
  stream: Stream;
  scrollWithstream = false;

  constructor(c: SavedConversation, messages: Message[]) {
    this.id = c.id;
    this.data = c;
    this.messages = messages;
    this.files = [];
    this.stream = newStream();

    // Initialize the stream with the conversation ID
    this.stream.init(this.id);
    this.stream.onDone = (finalContent: string) => this.onDone(finalContent);
    this.stream.onMessage((chunk) => this.handleMessage(chunk));
  }

  async onDone(finalContent: string) {
    // 1. Save to DB
    const savedAssistantMessage = await db.messages.create(this.id, {
      role: "assistant",
      content: finalContent,
    });
    // 2. Update Element to not streaming and set the id
    this.getAssistantMessageStreamEl()?.isDone(savedAssistantMessage.id);
    //3. scroll Anchor reset
    this.scrollWithstream = true;
  }

  getAssistantMessageStreamEl(): AssistantMessageElement | null {
    const all = document.querySelectorAll(`assistant-message#loadingMessage`);
    const messageEl = all[all.length - 1];

    if (!messageEl) {
      console.error("Could not find assistant message element to update");
      return null;
    }
    return messageEl as AssistantMessageElement;
  }

  scrollToEnd(force?: boolean) {
    if (!this.scrollWithstream && !force) return;
    requestAnimationFrame(() => {
      const scrollContainer = document.querySelector("#scrollContainer");
      if (scrollContainer) {
        scrollContainer.scrollTop = scrollContainer.scrollHeight;
      }
    });
  }

  handleMessage(chunk: string) {
    const messageEl = this.getAssistantMessageStreamEl();
    if (!messageEl) {
      console.error("cant handle message");
      return;
    }

    // Append the new content
    messageEl.appendContent(chunk);

    // Scroll to the bottom
    this.scrollToEnd();
  }

  getMessages() {
    return db.messages.getByConversation(this.id);
  }

  setLoading(isLoading: boolean) {
    if (isLoading) {
      return this.addMessageToDOM({
        role: "assistant",
        content: ". . .",
        id: "loadingMessage",
        timestamp: 0,
        conversationId: this.id,
      });
    }
  }

  async drawMessages(): Promise<void> {
    const messages = await db.messages.getByConversation(this.id);
    const rootEl =
      document.querySelector("#messageRoot") ||
      document.querySelector("#inner");

    if (!rootEl) {
      console.log("no root found for messages");
      return;
    }

    rootEl.innerHTML = "";

    // Create a promise that resolves after all messages are added to the DOM
    const appendPromises = messages.map((m) => this.addMessageToDOM(m));

    // Wait for all messages to be appended
    await Promise.all(appendPromises);

    // Return a promise that resolves on next frame when browser has rendered
    return new Promise((resolve) => requestAnimationFrame(() => resolve()));
  }

  async addMessageToDOM(message: Message) {
    const rootEl = document.querySelector("#messageRoot");
    if (!rootEl) return Promise.resolve();

    const div = document.createElement("div");
    if (message.role === "user")
      div.innerHTML = `<user-message>${message.content}</user-message>`;
    if (message.role === "assistant")
      div.innerHTML = `<assistant-message id="${message.id}">${message.content}</assistant-message>`;

    rootEl?.appendChild(div);

    // Return a promise to track when this message is added
    return Promise.resolve();
  }

  imageToBase64(file: File) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.readAsDataURL(file);
      reader.onload = () => resolve(reader.result);
      reader.onerror = (error) => reject(error);
    });
  }

  async handleUserInput(content?: string) {
    if (content.trim() === "") return; // Don't send empty content

    const scrollContainer = document.querySelector("#scrollContainer");

    this.scrollWithstream = true;
    // Use an arrow function to preserve 'this'
    // Use arrow function to preserve the 'this' context
    const scrollHandler = () => {
      if (scrollContainer) {
        // Only disable auto-scroll if user scrolls up or away from bottom
        const isAtBottom =
          scrollContainer.scrollHeight - scrollContainer.scrollTop <=
          scrollContainer.clientHeight + 50; // 50px threshold

        if (!isAtBottom) {
          this.scrollWithstream = false;
          scrollContainer.removeEventListener("scroll", scrollHandler);
        }
      }
    };

    scrollContainer?.addEventListener("scroll", scrollHandler);

    // Save message to db
    if (content) {
      const savedUserMessage = await db.messages.create(this.id, {
        role: "user",
        content,
      });
      this.addMessageToDOM(savedUserMessage);
    }
    // this.addMessageToDOM(assistantMessagePlaceholder);
    // this.messageIdForStream = assistantMessagePlaceholder.id;
    // this.setLoading(true);
    this.setLoading(true);

    // Get past messages
    const messages = await this.getMessages();
    const openMessages = messages.map((m: Message) => ({
      role: m.role,
      content: [
        {
          type: "text",
          text: "whats in this image?",
        },
        {
          type: "image_url",
          image_url: {
            url: "data:image/png;base64,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",
          },
        },
      ],
      id: m.id,
    })) as ChatCompletionMessageParam[];

    const contextMessages = openMessages.slice(-5);

    // get files
    const conv = await db.conversation.get(this.id);
    const hasFiles = !!conv?.files.map((f) => f.id).length;

    console.log({ hasFiles, c: conv });

    await this.stream.sendMessage(contextMessages, this.id, hasFiles);
  }
}

// async function getCompletion(
//     messages: ChatCompletionMessageParam[],
//     conversationId?: string,
//     hasFiles?: boolean,
// ): Promise<string> {
//   try {
//     const res = await axios.post(`/send-message`, {
//       messages,
//       conversationId,
//       hasFiles,
//     });
//     return res.data.content;
//   } catch (err) {
//     console.log(err);
//     return "Oops, something went wrong. We're sending a developer to fix it.";
//   }
// }

export { Conversation, getCompletion };

export async function uploadFile(file: File, conversationId: string) {
  if (!file) {
    throw new Error("missing param file");
  }
  if (!conversationId) {
    throw new Error("missing param conversationId");
  }
  const { data } = await axios.post(
    `/file/upload`,
    { file, conversationId },
    {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    },
  );
  const fileId = data.id;

  // Todo, maybe we're doing too much here
  const conv = await db.conversation.get(conversationId);
  if (!conv) return null;
  await db.conversation.update({
    ...conv,
    files: [...conv.files, { id: fileId, name: file.name }],
  });
  return fileId;
}

export async function removeFile(fileId: string, conversationId: string) {
  const { data } = await axios.post(
    `/file/delete`,
    { fileId, conversationId },
    {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    },
  );

  const conv = await db.conversation.get(conversationId);
  if (!conv) return;
  await db.conversation.update({
    ...conv,
    files: [...conv.files.filter((f) => f.id !== fileId)],
  });
}

export async function removeConversation(conversationId: string) {
  axios.post(
    `/conversation/delete`,
    { conversationId },
    {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    },
  );
  db.conversation.delete(conversationId);
}

export async function createInStorage() {
  const id = nanoid().replace(/-/gi, "");
  db.conversation.create(id);
  return id;
}

export async function createUserMessage(content: string, id: string) {
  db.messages.create(id, { role: "user", content });
}

async function getCompletion(
  messages: ChatCompletionMessageParam[],
  conversationId?: string,
  hasFiles?: boolean,
): Promise<string> {
  try {
    const res = await axios.post(`/send-message`, {
      messages,
      conversationId,
    });
    return res.data.content;
  } catch (err) {
    console.log(err);
    return "Nieuw gesprek";
  }
}

export async function initConversation(id: string) {
  const savedConversation = await db.conversation.get(id);
  window.history.replaceState(null, "", window.location.origin + `/c/${id}`);

  //  Todo: redirect to 404 page or something else
  if (!savedConversation) {
    return console.log("NO CONVERSATION FOUND FOR:", id);
  }

  if (!savedConversation) throw Error("failed saving conversation");
  const messages = await db.messages.getByConversation(id);

  const lastMessage = messages.slice(-1, messages.length)[0];
  const conversation = new Conversation(savedConversation, messages);
  await conversation.drawMessages();

  // Give the browser one more animation frame to finish rendering
  conversation.scrollToEnd(true);

  if (lastMessage?.role === "user") {
    await conversation.handleUserInput();
    // try {
    //   conversation.setLoading(true);
    //   const conv = await db.conversation.get(id);
    //   const hasFiles = !!conv?.files.map((f) => f.id).length;
    //   const openMessages = messages.map(
    //     (m) =>
    //       ({
    //         role: m.role,
    //         content: m.content,
    //       }) as ChatCompletionUserMessageParam,
    //   );
    //   const answer = await getCompletion(openMessages, id, hasFiles);
    //   const savedAssistantMessage = await db.messages.create(conversation.id, {
    //     role: "assistant",
    //     content: answer,
    //   });
    //   conversation.setLoading(false);
    //   await conversation.addMessageToDOM(savedAssistantMessage);
    // } catch (err) {
    //   console.log({ err });
    //   const savedAssistantMessage = await db.messages.create(conversation.id, {
    //     role: "assistant",
    //     content: "Oeps, er ging iets mis.",
    //   });
    //   conversation.setLoading(false);
    //   await conversation.addMessageToDOM(savedAssistantMessage);
    // }
  }

  if (messages.length === 1) {
    const content = messages[0].content;
    const prompt = `
            I want you to summarize the message below into 1 short sentence so it can serve as the title of the conversation the message is opening.
            Your response will be directly serving as the title, so please just respond with the title and nohing else
            message:
            ${content}
        `;
    const title = await getCompletion([{ role: "user", content: prompt }]);
    db.conversation.update({ ...savedConversation, title });

    //   Test
  }
  (window as any).goChat.conversation = conversation;

  // await conversation.drawMessages();
}
