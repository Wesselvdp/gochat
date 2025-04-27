import { LitElement, unsafeCSS } from "lit";
import globalStyles from "../styles.scss?inline";

export class TailwindElement extends LitElement {
  static styles = [unsafeCSS(globalStyles)];
}
