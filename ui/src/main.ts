import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import './userWorker';
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api';
import { conf, language} from "monaco-editor/esm/vs/basic-languages/sql/sql";

monaco.languages.register({ id: "sql" });
monaco.languages.setMonarchTokensProvider("sql", language);

createApp(App).mount('#app')
