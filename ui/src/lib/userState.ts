import { ref } from "vue";

export interface User {
  firstName: string;
  lastName: string;
  email: string;
}

export const currentUser = ref<User | null>(null);
