<script setup lang="ts">
import { ref, watch, watchEffect } from "vue";
import {APIError, doAPIRequest, formatError} from "../lib/api";
import { Dataset, useDatasets, Query } from "../lib/datasetState";
import Confetti from "./Confetti.vue";
import Editor from "./Editor.vue";

defineEmits(["goBack"]);
const props = defineProps<{
  datasetId: string;
  queryId: string;
}>();

const {datasets, refresh: refreshDatasets} = useDatasets();

const input = ref("");
const resultJSON = ref("");
const message = ref("");
const messageType = ref<"success" | "error" | null>(null);
const loading = ref(false);

const confettiRef = ref<typeof Confetti>();

const dataset = ref<Dataset | null>(null);
const query = ref<Query | null>(null);
watchEffect(() => {
  if (datasets === null) {
    return;
  }
  dataset.value = datasets.find(x => x.id === props.datasetId)!;
  query.value = dataset.value.queries.find(x => x.id === props.queryId)!;
});

async function doQuery() {
  if (input.value.length === 0) {
    return;
  }
  loading.value = true;
  try {
    const result = await doAPIRequest(
      "POST",
      `/dataset/${props.datasetId}/query`,
      200,
      {
        statement: input.value,
      }
    );
    resultJSON.value = JSON.stringify(result, null, 2);
  } catch (e) {
    message.value = formatError(e);
  } finally {
    loading.value = false;
  }
}

async function doCheck() {
  if (input.value.length === 0) {
    return;
  }
  loading.value = true;
  try {
    const result = await doAPIRequest(
      "POST",
      `/dataset/${props.datasetId}/${props.queryId}/submitAnswer`,
      200,
      {
        statement: input.value,
      }
    ) as {points: number};
    message.value = `Congratulations, that was the correct query! You have received ${result.points} points.`;
    messageType.value = "success"; // if the API didn't error we know it's correct
    confettiRef.value?.fire({});
    refreshDatasets();
  } catch (e) {
    message.value = formatError(e);
    messageType.value = "error";
  } finally {
    loading.value = false;
  }
}

async function getHint() {
  try {
    loading.value = true;
    const result = await doAPIRequest(
        "POST",
        `/dataset/${props.datasetId}/${props.queryId}/useHint`,
        200,
        {}
    ) as Query;
    const dsIdx = datasets!.findIndex(x => x.id === props.datasetId);
    const qIdx = datasets![dsIdx].queries.findIndex(x => x.id === props.queryId);
    datasets![dsIdx].queries[qIdx] = result;
  } catch (e) {
    message.value = formatError(e);
    messageType.value = "error";
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div v-if="dataset === null || query === null">
    <h1>Loading, please wait...</h1>
  </div>
  <div v-else>
    <h1>{{ query.name }}</h1>
    <button @click="$emit('goBack')">Go Back</button>
    <p class="desc">{{ query.challenge }}</p>

    <div v-if="query.hints !== null">
      <button v-if="query.hints.length < query.numHints" class="small" @click="getHint">Stuck? Get a hint!</button>
      <ul>
        <li v-for="hint  in query.hints">{{hint}}</li>
      </ul>
    </div>

    <Editor
      v-model="input"
      language="sql"
    ></Editor>
    <div>
      <button :disabled="loading" @click="doQuery">Run Query</button>
      <button :disabled="loading" @click="doCheck" class="check">
        Check Answer
      </button>
    </div>
    <div v-if="message" class="message" :class="messageType">{{ message }}</div>
    <Editor v-if="resultJSON" v-model="resultJSON" language="json" readonly></Editor>
  </div>
  <confetti ref="confettiRef"></confetti>
</template>

<style scoped>
.desc {
  max-width: 48rem;
}
.check {
  background-color: #104f5f;
  color: white;
  font-weight: bold;
}
.check[disabled] {
  background-color: #495057;
  color: white;
  font-weight: bold;
}
.message {
  max-width: 80vw;
  word-wrap: break-word;
}
.error {
  background-color: #6b0700;
  color: #fafafa;
}
.success {
  background-color: #006b22;
  color: #fafafa;
}
.small {
  font-size: 80%;
}
</style>
