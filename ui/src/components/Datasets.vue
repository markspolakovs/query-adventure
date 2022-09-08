<script setup lang="ts">
import {onMounted, ref} from "vue";
import { doAPIRequest } from "../lib/api";
import {useDatasets} from "../lib/datasetState";
import Challenge from "./Challenge.vue";

const active = ref<null | [string, string]>(null);

const datasets = useDatasets();
onMounted(datasets.refresh);
</script>

<template>
  <Challenge
    v-if="active !== null"
    :dataset-id="active[0]"
    :query-id="active[1]"
    @go-back="active = null"
  />
  <div v-else class="list">
    <h2>Datasets</h2>
    <b v-if="datasets.datasets === null">Loading, please wait...</b>
    <div v-if="datasets.datasets !== null" v-for="ds in datasets.datasets" :key="ds.id">
      <h3>{{ ds.name }}</h3>
      <p>{{ ds.description }}</p>
      <p>
        <b>Challenges:</b>
        <ul>
            <li v-for="q in ds.queries" :key="q.id">
                <button
                    @click="active = [ds.id, q.id]">
                  <del v-if="q.complete">{{ q.name }}</del>
                  <span v-else>{{ q.name }}</span>
                </button>
            </li>
        </ul>
      </p>
    </div>
  </div>
</template>

<style scoped>
.list {
  max-width: 64rem;
}
ul li {
    list-style-type: none;
}
</style>
