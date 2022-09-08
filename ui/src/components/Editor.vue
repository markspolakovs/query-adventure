<template>
  <div ref="editorRef" class="editor"></div>
</template>

<script setup lang="ts">
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api';
import {onMounted, ref, watch} from "vue";
import {editor} from "monaco-editor/esm/vs/editor/editor.api";
import IStandaloneCodeEditor = editor.IStandaloneCodeEditor;
const props = defineProps<{
  modelValue: string,
  language: string
  readonly?: boolean
}>()
const emits = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>();

const editorRef = ref<HTMLDivElement>();

let monacoEditor: IStandaloneCodeEditor;
onMounted(() => {
  monacoEditor = monaco.editor.create(editorRef.value!, {
    value: props.modelValue,
    language: props.language,
    readOnly: props.readonly
  })
  let cursorState: monaco.Selection[] | null;
  monacoEditor.getModel()!.onDidChangeContent(e => {
    cursorState = monacoEditor.getSelections();
    emits("update:modelValue", monacoEditor.getValue());
  });
  watch(() => props.modelValue, v => {
    const model = monacoEditor.getModel()!;
    model.pushEditOperations([], [{
      range: model.getFullModelRange(),
      text: v
    }], undefined as any);
  });
});
</script>

<style scoped>
.editor {
  text-align: initial;
  width: 100%;
  min-width: 48rem;
  min-height: 24rem;
  box-shadow: 1px 1px 3px rgba(0, 0, 0, 0.4);
}
</style>