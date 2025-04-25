import "./styles.scss";
// import { initConversation } from "./conversation";
import "./components";
import db, { SavedConversation } from "./db";
import "./htmx-openai";
// Make the function available on the window object
import { initSentry } from "./sentry";
import "./components";
import "./svg/icons";

import htmx from "htmx.org";

// Initialize Sentry before your app
initSentry();

(async () => {
  // recentConversations.init()
})();
