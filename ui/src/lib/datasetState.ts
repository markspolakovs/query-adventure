import { ref } from "vue";

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
  complete: boolean;
}

export const datasets = ref<Dataset[] | null>(null);
