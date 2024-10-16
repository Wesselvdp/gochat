

declare const htmx: any;



htmx.defineExtension("openai", {
  onEvent: async function (name: string, evt: CustomEvent) {
    if (name === "htmx:beforeSend") {
      evt.preventDefault();
      // evt.detail.headers['Content-Type'] = 'application/json';
      const messages = await (window as any).goChat.conversation.getMessages();
      htmx.ajax("POST", `send-message`, {
        messages
      });
      // evt.detail.xhr.send(JSON.stringify(jsonData));
    }
  },
});

