import { html, PropertyValues } from "lit";

import { customElement, property, state } from "lit/decorators.js";

import { TailwindElement } from "./albertElement";
import { ChatService } from "../application/ChatService";
import { DexieThreadRepository } from "../infrastructure/persistence/dexieThreadRepository";
import { ModelParams, Thread } from "../domain";

type ModelBehaviour = {
  id: string;
  label: string;
  temperature: number;
  top_p: number;
  description: string;
};

const modelBehaviours: ModelBehaviour[] = [
  {
    id: "precision",
    label: "Precisie Modus",
    temperature: 0.0,
    top_p: 0.7,
    description:
      "Gebruik wanneer je consistente, deterministische resultaten nodig hebt met minimale variatie. Perfect voor reproduceerbare data-analyse, feitelijke zoekopdrachten of gestructureerde gegevensextractie.",
  },
  {
    id: "balanced",
    label: "Standaard Onderzoeker",
    temperature: 0.3,
    top_p: 0.8,
    description:
      "Behoudt hoge nauwkeurigheid met een lichte variabiliteit. Ideaal voor het samenvatten van onderzoeksresultaten, het genereren van inzichten uit data of het creëren van goed gestructureerde rapporten.",
  },
  {
    id: "creative",
    label: "Ideeën Verkenner",
    temperature: 0.7,
    top_p: 0.9,
    description:
      "Stimuleert meer diverse perspectieven met behoud van relevantie. Uitstekend voor het brainstormen over onderzoeksrichtingen, het verkennen van alternatieve interpretaties van data of het ontdekken van nieuwe verbanden.",
  },
  {
    id: "exploratory",
    label: "Ontdekking Maximaal",
    temperature: 1.0,
    top_p: 1.0,
    description:
      "Produceert het breedste bereik aan mogelijke outputs. Nuttig voor divergent denken, het ontdekken van onverwachte patronen in data of het genereren van meerdere hypotheses om te testen.",
  },
  {
    id: "technical",
    label: "Rapport Generator",
    temperature: 0.2,
    top_p: 0.85,
    description:
      "Geoptimaliseerd voor het genereren van technische inhoud die nauwkeurigheid behoudt terwijl herhalend taalgebruik wordt vermeden. Uitstekend voor het opstellen van methodologiesecties of technische documentatie.",
  },
];

@customElement("dialog-message-settings")
export class userInputForm extends TailwindElement {
  async connectedCallback() {
    super.connectedCallback();

    // const modelParams = await this.chatService.getModelParams(this.threadId);
    const modelParams = this.modelParams;

    if (modelParams) {
      this.modelBehaviour =
        modelBehaviours.find(
          (profile) =>
            profile.temperature === modelParams.temperature &&
            profile.top_p === modelParams.top_p,
        ) || modelBehaviours[0];
    }

    if (!modelParams) this.modelBehaviour = modelBehaviours[1];

    const thread = await this.chatService.getThread(this.threadId);
    this.thread = thread;
  }

  private chatService: ChatService;

  constructor() {
    super();
    const threadRepository = new DexieThreadRepository();
    this.chatService = new ChatService(threadRepository);
  }

  @property()
  modelParams: ModelParams = {};

  @state()
  modelBehaviour: ModelBehaviour | undefined = undefined;
  @property()
  setModelParams: (modelParams: ModelParams) => void = () => {};
  @property()
  threadId: string = "";

  @state()
  thread: Thread | undefined = undefined;

  protected updated(_changedProperties: PropertyValues) {
    super.updated(_changedProperties);
    if (_changedProperties.has("modelBehaviour")) {
      this.setModelParams({
        temperature: this.modelBehaviour?.temperature,
        top_p: this.modelBehaviour?.top_p,
      });
    }
  }

  render() {
    return html`
      <dialog-content id="dialogContent">
        <dialog-header>
          <dialog-title>
            <div>Gespreksopties</div>
          </dialog-title>

          <dialog-description></dialog-description>
        </dialog-header>

        <div class=" ">
          <div class="mb-8">
            <torgon-input
              label="Gespreksnaam"
              @input-blur="${(e: any) => {
                this.chatService.renameThread(this.threadId, e.detail.value);
              }}"
              .value="${this.thread?.title}"
            ></torgon-input>
          </div>

          <p class="font-semibold leading-none tracking-tight mb-2">
            Creativiteit
          </p>
          <div class="flex w-full gap-8">
            <div class="mb-2 w-1/3 flex flex-row flex-wrap gap-2">
              ${modelBehaviours.map(
                (option) => html`
                  <torgon-button
                    class="w-full"
                    variant="${this.modelBehaviour.id === option.id
                      ? "secondary"
                      : "outline"}"
                    @click="${() => {
                      this.modelBehaviour = option;
                    }}"
                    >${option.label}
                  </torgon-button>
                `,
              )}
            </div>
            <div class="w-2/3">
              <p class="text-sm mb-4">${this.modelBehaviour?.description}</p>
              <div class="text-xs opacity-50">
                <p>temperatuur: ${this.modelBehaviour?.temperature}</p>
                <p>top_p: ${this.modelBehaviour?.top_p}</p>
              </div>
            </div>
          </div>
        </div>

        <dialog-footer></dialog-footer>
      </dialog-content>
    `;
  }
}
