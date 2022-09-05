<script setup lang="ts">
import { ref, watch, watchEffect } from "vue";
import { APIError, doAPIRequest } from "../lib/api";
import { Dataset, datasets, Query } from "../lib/datasetState";

defineEmits(["goBack"]);
const props = defineProps<{
  datasetIdx: number;
  queryIdx: number;
}>();

const input = ref("");
const status = ref("");
const resultJSON = ref("");
const resultType = ref<"success" | "error" | null>(null);
const loading = ref(false);
const hintIndex = ref(0);

const dataset = ref<Dataset | null>(null);
const query = ref<Query | null>(null);
watchEffect(() => {
  if (datasets.value === null) {
    return;
  }
  dataset.value = datasets.value[props.datasetIdx];
  query.value = dataset.value.queries[props.queryIdx];
});

async function doQuery() {
  if (input.value.length === 0) {
    return;
  }
  loading.value = true;
  try {
    const result = await doAPIRequest(
      "POST",
      `/dataset/${props.datasetIdx}/query`,
      200,
      {
        statement: input.value,
      }
    );
    resultJSON.value = JSON.stringify(result, null, 2);
    resultType.value = null;
  } catch (e) {
    if (e instanceof APIError) {
      resultJSON.value = e.message;
    } else if (e instanceof Error) {
      resultJSON.value = e.toString();
    } else {
      resultJSON.value = String(e);
    }
    resultType.value = "error";
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
      `/dataset/${props.datasetIdx}/${props.queryIdx}/submitAnswer`,
      200,
      {
        statement: input.value,
      }
    );
    resultJSON.value = JSON.stringify(result, null, 2); // FIXME this will change
    resultType.value = "success"; // if the API didn't error we know it's correct
  } catch (e) {
    if (e instanceof APIError) {
      resultJSON.value = e.message;
    } else if (e instanceof Error) {
      resultJSON.value = e.toString();
    } else {
      resultJSON.value = String(e);
    }
    resultType.value = "error";
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
    <p>{{ query.challenge }}</p>

    <div v-if="query.hints !== null">
      <button v-if="hintIndex < query.hints.length" class="small" @click="hintIndex++">Stuck? Get a hint!</button>
      <ul>
        <li v-for="hint  in query.hints.slice(0, hintIndex)">{{hint}}</li>
      </ul>
    </div>

    <textarea
      v-model="input"
      placeholder="SELECT * FROM ..."
      rows="5"
      cols="80"
    ></textarea>
    <div>
      <button :disabled="loading" @click="doQuery">Run Query</button>
      <button :disabled="loading" @click="doCheck" class="check">
        Check Answer
      </button>
    </div>
    <p>{{ status }}</p>
    <div v-if="resultJSON" class="output" :class="resultType">{{ resultJSON }}</div>
  </div>
</template>

<style scoped>
.check {
  background-color: #104f5f;
  color: white;
  font-weight: bold;
}
.output {
  /*overflow: auto;*/
  word-wrap: break-word;
  text-align: start;
  padding: 0.4rem;
  font-family: monospace;
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
