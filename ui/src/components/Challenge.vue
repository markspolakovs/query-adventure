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
const loading = ref(false);

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
  } catch (e) {
    if (e instanceof APIError) {
      resultJSON.value = e.message;
    } else if (e instanceof Error) {
      resultJSON.value = e.toString();
    } else {
      resultJSON.value = String(e);
    }
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
    resultJSON.value = JSON.stringify(result, null, 2);
  } catch (e) {
    if (e instanceof APIError) {
      resultJSON.value = JSON.stringify(JSON.parse(e.message), null, 2);
    } else if (e instanceof Error) {
      resultJSON.value = e.toString();
    } else {
      resultJSON.value = String(e);
    }
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
    <pre v-if="resultJSON" class="output">{{ resultJSON }}</pre>
  </div>
</template>

<style scoped>
.check {
  background-color: #104f5f;
  color: white;
  font-weight: bold;
}
.output {
  overflow: auto;
  overflow-wrap: break-word;
  text-align: start;
}
</style>
