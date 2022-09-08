import { ref } from "vue";
import {defineStore} from "pinia";
import {doAPIRequest, formatError} from "./api";

// These match rest.apiDataset/apiQuery, *not* data.Dataset/Query

export interface Dataset {
  id: string;
  name: string;
  description: string;
  queries: Query[];
}

export interface Query {
  id: string;
  name: string;
  challenge: string;
  points: number;
  hints: string[] | null;
  numHints: number;
  complete: boolean;
}

export const useDatasets = defineStore("datasets", {
  state: () => ({datasets: null as Dataset[] | null, error: null as string | null}),
  actions: {
    async refresh() {
      try {
        this.datasets = await doAPIRequest("GET", "/datasets", 200) as Dataset[]
      } catch (e) {
        this.error = formatError(e);
      }
    }
  }
});
