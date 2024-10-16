import "./styles.scss";
import {initConversation} from "./conversation";
import './components'
import db from "./db";
import './htmx-openai'
// Make the function available on the window object



const recentConversations = {
    init: async function() {
        const rootEl = document.querySelector('#recentConversationsRoot');
        if (!rootEl) return;
        rootEl.innerHTML = "";
        const conversations =  await db.conversation.list()
        conversations.map((c, i) => {
            if(i > 6) return;

            const div = document.createElement("div");
            div.innerHTML = `<recent-conversation id="${c.id}" title="${c.title}"></recent-conversation>`;
            rootEl?.appendChild(div)
        })
    },
};
(window as any).goChat = { initConversation, recentConversations };

(async () => {
   recentConversations.init()

} )()

// document.body.addEventListener('htmx:load', function(event) {
//     if (typeof initChat === 'function') {
//         initChat();
//     }
// });