<script setup lang="ts">
import { ref } from "vue";
import { doAPIRequest } from "../lib/api";
import { Dataset, datasets } from "../lib/datasetState";
import Challenge from "./Challenge.vue";

const active = ref<null | [string, string]>(null);

(async function () {
  datasets.value = (await doAPIRequest("GET", "/datasets", 200)) as Dataset[];
})();
</script>

<template>
  <Challenge
    v-if="active !== null"
    :dataset-id="active[0]"
    :query-id="active[1]"
    @go-back="active = null"
  />
  <div v-else>
    <h2>Datasets</h2>
    <b v-if="datasets === null">Loading, please wait...</b>
    <div v-if="datasets !== null" v-for="ds in datasets" :key="ds.id">
      <h3>{{ ds.name }}</h3>
      <p>{{ ds.description }}</p>
      <p>
        <b>Challenges:</b>
        <ul>
            <li v-for="q in ds.queries" :key="q.id">
                <button
                    @click="active = [ds.id, q.id]">
                {{ q.name }}
                </button>
            </li>
        </ul>
      </p>
    </div>
  </div>
</template>

<style scoped>
    ul li {
        list-style-type: none;
    }
</style>
