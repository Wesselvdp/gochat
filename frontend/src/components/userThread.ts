import { LitElement, css, html, unsafeCSS, PropertyValues } from "lit";
import { customElement, property, state } from "lit/decorators.js";
import globalStyles from "../styles.scss?inline";
import { Message, MessageStatus, messageStatusMap } from "../domain";
import { DexieThreadRepository } from "../infrastructure/persistence/dexieThreadRepository";
import { ChatService } from "../application/ChatService";
import {
  createStreamService,
  getStreamService,
} from "../application/StreamService";
import { AssistantMessageElement } from "./assistantMessage";

@customElement("user-thread")
export class UserThread extends LitElement {
  static styles = [unsafeCSS(globalStyles)];

  @property() id = "";
  @property() title = "";

  @state() messages: Message[] = [];

  statusesToShow: MessageStatus[] = [
    messageStatusMap.DONE,
    messageStatusMap.ERROR,
    messageStatusMap.STREAMING,
  ];

  private chatService: ChatService;
  // Track the last render time to prevent too frequent updates
  private lastRenderTime = 0;
  private renderDebounceTimeout: number | null = null;
  private messageSubscription: any = null;

  constructor() {
    super();
    const threadRepository = new DexieThreadRepository();
    this.chatService = new ChatService(threadRepository);
  }

  currentStreamingMessageElement: AssistantMessageElement | null = null;

  // Handle a new chunk arriving
  handleNewChunk = (chunk: string) => {
    if (!chunk) return;

    // If we don't have a reference to the streaming message element, find it
    if (!this.currentStreamingMessageElement) {
      if (!this.shadowRoot) {
        console.error("Shadow root is not available");
        return;
      }
      const streamingElements =
        this.shadowRoot.querySelectorAll("assistant-message") || [];
      const last = streamingElements[streamingElements.length - 1];

      console.log("Found streaming element:", { last });
      if (!last) {
        console.warn("No streaming element found to append chunk to");
        return;
      }

      this.currentStreamingMessageElement = last as AssistantMessageElement;
    }

    // Add the chunk directly to the element
    if (this.currentStreamingMessageElement) {
      this.currentStreamingMessageElement.appendContent(chunk);
    } else {
      console.error("No streaming message element to append to");
    }
  };

  protected update(changedProperties: PropertyValues) {
    super.update(changedProperties);
    if (changedProperties.has("messages")) {
      this.verifyThread();
    }
  }

  async verifyThread() {
    const thread = await this.chatService.getThread(this.id);
    if (!thread) {
      console.error("Thread not found");
      return;
    }
    const isNewThread = thread.title === "New Thread";
    const content = this.messages[0]?.content || "";

    if (!isNewThread || !content) return;
    this.chatService.generateThreadName(this.id, content);
  }

  async connectedCallback() {
    super.connectedCallback();
    // Initialize the stream service for this thread
    createStreamService(this.id, this.handleNewChunk);
    // Use a more controlled approach to subscribing to message updates
    this.messageSubscription = this.chatService.subscribeToMessages(
      this.id,
      (messages) => {
        const filteredMessages = messages.filter(
          (m) => m.status && this.statusesToShow.includes(m.status),
        );

        // Only update if messages have actually changed
        if (this.messagesChanged(this.messages, filteredMessages)) {
          // Debounce updates during streaming to prevent too many DOM updates
          if (this.renderDebounceTimeout) {
            clearTimeout(this.renderDebounceTimeout);
          }

          this.renderDebounceTimeout = window.setTimeout(() => {
            // Check for status changes from streaming to done
            const hasStreamingStatusChange = this.messages.some(
              (oldMsg) =>
                oldMsg.status === messageStatusMap.STREAMING &&
                filteredMessages.find(
                  (newMsg) =>
                    newMsg.id === oldMsg.id &&
                    newMsg.status === messageStatusMap.DONE,
                ),
            );

            if (hasStreamingStatusChange) {
              // Reset the current streaming element reference when a message completes
              this.currentStreamingMessageElement = null;
            }

            this.messages = filteredMessages;
            this.renderDebounceTimeout = null;
          }, 50); // Small debounce time to batch rapid updates
        }
      },
    );
  }

  // Helper to compare message arrays
  messagesChanged(oldMessages: Message[], newMessages: Message[]): boolean {
    if (oldMessages.length !== newMessages.length) return true;

    for (let i = 0; i < oldMessages.length; i++) {
      if (
        oldMessages[i].content !== newMessages[i].content ||
        oldMessages[i].status !== newMessages[i].status ||
        oldMessages[i].id !== newMessages[i].id
      ) {
        return true;
      }
    }

    return false;
  }

  // Fix the message rendering function to be consistent
  messageRenderFunc = (message: Message) => {
    if (message.role === "user") {
      // IMPORTANT: Pass content as child content, not attribute
      return html`<div>
        <user-message .content=${message.content}></user-message>
      </div>`;
    }

    if (message.role === "assistant") {
      // Use key attribute to help Lit track each message uniquely
      return html`<assistant-message
        .markdownContent=${message.content}
        .id=${message.id}
        .streaming=${message.status === messageStatusMap.STREAMING}
      ></assistant-message>`;
    }

    return html``; // Return empty fragment for unexpected roles
  };

  render() {
    // Use the map with a key function to help Lit update more efficiently
    return html`
      ${this.messages.map(
        (message) =>
          html`<div key="${message.id}">
            ${this.messageRenderFunc(message)}
          </div>`,
      )}
    `;
  }

  disconnectedCallback() {
    // Clean up any pending operations
    if (this.renderDebounceTimeout) {
      clearTimeout(this.renderDebounceTimeout);
    }

    // Clean up message subscription
    if (this.messageSubscription) {
      this.messageSubscription(); // Call the unsubscribe function
    }

    // Close the stream when the component is disconnected
    const streamService = getStreamService();
    if (streamService) {
      streamService.closeStream();
    }

    console.log("disconnectedCallback");

    super.disconnectedCallback();
  }
}
