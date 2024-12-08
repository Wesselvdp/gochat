import {html, css, LitElement, unsafeCSS} from 'lit';
import {customElement} from "lit/decorators.js";
import globalStyles from "../styles.scss?inline";

@customElement('icon-paperclip')
class iconPaperclip extends LitElement {
    static styles = [
        unsafeCSS(globalStyles),
        css``];

    render() {
        return html`
            <div class="flex">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round" class="feather feather-paperclip"><path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48"></path></svg>
            </div>
    `;
    }
}

@customElement('icon-remove')
class iconRemove extends LitElement {
    static styles = [
        unsafeCSS(globalStyles),
        css``];

    render() {
        return html`
            <div class="flex">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-circle-x"><circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/></svg>
            </div>
    `;
    }
}