import { ref } from "vue";
import { doAPIRequest } from "./api";

export interface Dataset {
  name: string;
  description: string;
  queries: Query[];
}

export interface Query {
  name: string;
  challenge: string;
  points: number;
}

export const datasets = ref<Dataset[] | null>(null);
