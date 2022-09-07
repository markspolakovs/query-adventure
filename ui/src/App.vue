<script setup lang="ts">
import AuthGate from "./components/AuthGate.vue";
import Datasets from "./components/datasets.vue";
import { currentUser } from "./lib/userState";
import {ref} from "vue";
import Scoreboard from "./components/Scoreboard.vue";

const path = ref(window.location.pathname);
</script>

<template>
  <AuthGate>
    <div v-if="currentUser && path !== '/scoreboard'">
      <b>Hello, {{ currentUser.firstName }}</b>
    </div>
    <Scoreboard v-if="path === '/scoreboard'" />
    <Datasets v-else />
  </AuthGate>
</template>

<style scoped></style>
