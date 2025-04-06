// conversation.ts
import db, {Message, SavedConversation} from "./db";
import axios from "axios";
import {nanoid} from "nanoid";
import {ChatCompletionMessageParam} from "openai/src/resources/chat/completions";
import {Stream, newStream} from "./stream";
import {AssistantMessageElement} from "./components/assistantMessage";

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
        const all = document.querySelectorAll(
            `assistant-message#loadingMessage`
        )
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

    async handleUserInput(content?: string) {
        // if (content.trim() === "") return; // Don't send empty content

        const scrollContainer = document.querySelector("#scrollContainer");

        this.scrollWithstream = true;
        // Use an arrow function to preserve 'this'
        // Use arrow function to preserve the 'this' context
        const scrollHandler = () => {
            console.log("scroll handler!!");
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
            content: m.content,
            id: m.id,
        })) as ChatCompletionMessageParam[];
        const contextMessages = openMessages.slice(-5);

        // get files
        const conv = await db.conversation.get(this.id);
        const hasFiles = !!conv?.files.map((f) => f.id).length;

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

export {Conversation, getCompletion};

export async function uploadFile(file: File, conversationId: string) {
    if (!file) {
        throw new Error("missing param file");
    }
    if (!conversationId) {
        throw new Error("missing param conversationId");
    }
    const {data} = await axios.post(
        `/file/upload`,
        {file, conversationId},
        {
            headers: {
                "Content-Type": "multipart/form-data",
            },
        }
    );
    const fileId = data.id;

    // Todo, maybe we're doing too much here
    const conv = await db.conversation.get(conversationId);
    if (!conv) return null;
    await db.conversation.update({
        ...conv,
        files: [...conv.files, {id: fileId, name: file.name}],
    });
    return fileId;
}

export async function removeFile(fileId: string, conversationId: string) {
    const {data} = await axios.post(
        `/file/delete`,
        {fileId, conversationId},
        {
            headers: {
                "Content-Type": "multipart/form-data",
            },
        }
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
        {conversationId},
        {
            headers: {
                "Content-Type": "multipart/form-data",
            },
        }
    );
    db.conversation.delete(conversationId);
}

export async function createInStorage() {
    const id = nanoid().replace(/-/gi, "");
    db.conversation.create(id);
    return id;
}

export async function createUserMessage(content: string, id: string) {
    db.messages.create(id, {role: "user", content});
}

async function getCompletion(
    messages: ChatCompletionMessageParam[],
    conversationId?: string,
    hasFiles?: boolean
): Promise<string> {
    try {
        const res = await axios.post(`/send-message`, {
            messages,
            conversationId,
            hasFiles,
        });
        return res.data.content;
    } catch (err) {
        console.log(err);
        return "Oeps, er is iets mis. We sturen er een ontwikkelaar op af";
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
            
            message:
            ${content}
        `;
        const title = await getCompletion([{role: "user", content: prompt}]);
        db.conversation.update({...savedConversation, title});
    }
    (window as any).goChat.conversation = conversation;

    // await conversation.drawMessages();
}
