<script setup lang="ts">
import { cpp } from "@codemirror/lang-cpp";
import { css } from "@codemirror/lang-css";
import { go } from "@codemirror/lang-go";
import { html } from "@codemirror/lang-html";
import { java } from "@codemirror/lang-java";
import { javascript } from "@codemirror/lang-javascript";
import { json } from "@codemirror/lang-json";
import { markdown } from "@codemirror/lang-markdown";
import { php } from "@codemirror/lang-php";
import { python } from "@codemirror/lang-python";
import { rust } from "@codemirror/lang-rust";
import { sql } from "@codemirror/lang-sql";
import { yaml } from "@codemirror/lang-yaml";
import {
  defaultHighlightStyle,
  syntaxHighlighting,
  type LanguageSupport,
} from "@codemirror/language";
import { Compartment, EditorState } from "@codemirror/state";
import { basicSetup, EditorView } from "codemirror";

const props = withDefaults(
  defineProps<{
    modelValue: string;
    language?: string;
    readOnly?: boolean;
    label?: string;
  }>(),
  { language: "text", readOnly: false, label: "代码编辑器" },
);

const emit = defineEmits<{ "update:modelValue": [value: string] }>();
const host = ref<HTMLElement>();
let editor: EditorView | undefined;
const languageCompartment = new Compartment();
const editableCompartment = new Compartment();

function languageExtension(language: string): LanguageSupport | readonly [] {
  switch (language) {
    case "cpp": return cpp();
    case "css": return css();
    case "go": return go();
    case "html":
    case "vue": return html();
    case "java": return java();
    case "javascript": return javascript({ jsx: true });
    case "typescript": return javascript({ jsx: true, typescript: true });
    case "json": return json();
    case "markdown": return markdown();
    case "php": return php();
    case "python": return python();
    case "rust": return rust();
    case "sql": return sql();
    case "yaml": return yaml();
    default: return [];
  }
}

const pasteTheme = EditorView.theme({
  "&": {
    height: "100%",
    backgroundColor: "var(--paste-editor)",
    color: "var(--paste-ink)",
    fontSize: "14px",
  },
  ".cm-scroller": {
    fontFamily: '"SFMono-Regular", Consolas, "Liberation Mono", monospace',
    lineHeight: "1.72",
    overflow: "auto",
  },
  ".cm-content": { padding: "18px 0 30vh", caretColor: "var(--paste-blue)" },
  ".cm-line": { padding: "0 20px 0 10px" },
  ".cm-gutters": {
    backgroundColor: "var(--paste-editor)",
    color: "var(--paste-ink-dim)",
    border: "0",
    paddingLeft: "8px",
  },
  ".cm-activeLine, .cm-activeLineGutter": {
    backgroundColor: "var(--paste-editor-active)",
  },
  ".cm-selectionBackground, &.cm-focused .cm-selectionBackground, ::selection": {
    backgroundColor: "var(--paste-selection) !important",
  },
  "&.cm-focused": { outline: "none" },
  ".cm-cursor": { borderLeftColor: "var(--paste-blue)" },
});

onMounted(() => {
  if (!host.value) return;
  editor = new EditorView({
    parent: host.value,
    state: EditorState.create({
      doc: props.modelValue,
      extensions: [
        basicSetup,
        pasteTheme,
        syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
        EditorView.contentAttributes.of({ "aria-label": props.label }),
        languageCompartment.of(languageExtension(props.language)),
        editableCompartment.of(EditorView.editable.of(!props.readOnly)),
        EditorState.readOnly.of(props.readOnly),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) emit("update:modelValue", update.state.doc.toString());
        }),
      ],
    }),
  });
});

watch(
  () => props.modelValue,
  (value) => {
    if (!editor || value === editor.state.doc.toString()) return;
    editor.dispatch({ changes: { from: 0, to: editor.state.doc.length, insert: value } });
  },
);

watch(
  () => props.language,
  (value) => editor?.dispatch({ effects: languageCompartment.reconfigure(languageExtension(value)) }),
);

watch(
  () => props.readOnly,
  (value) => editor?.dispatch({ effects: editableCompartment.reconfigure(EditorView.editable.of(!value)) }),
);

watch(
  () => props.label,
  (value) => editor?.contentDOM.setAttribute("aria-label", value),
);

onBeforeUnmount(() => editor?.destroy());
</script>

<template>
  <div ref="host" class="paste-code-editor" role="region" :aria-label="label" />
</template>
