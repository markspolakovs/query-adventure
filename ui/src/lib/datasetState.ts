import { ref } from "vue";

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
}

export const datasets = ref<Dataset[] | null>(null);
