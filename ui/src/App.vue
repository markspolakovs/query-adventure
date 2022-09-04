<script setup lang="ts">
import AuthGate from "./components/AuthGate.vue";
import Datasets from "./components/datasets.vue";
import { currentUser } from "./lib/userState";
</script>

<template>
  <AuthGate>
    <div v-if="currentUser">
      <b>Hello, {{ currentUser.firstName }}</b>
    </div>
    <Datasets />
  </AuthGate>
</template>

<style scoped></style>
