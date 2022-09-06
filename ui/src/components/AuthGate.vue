<script setup lang="ts">
// This starter template is using Vue 3 <script setup> SFCs
// Check out https://vuejs.org/api/sfc-script-setup.html#script-setup
import { ref } from "vue";
import { currentUser, User } from "../lib/userState";
import { doAPIRequest, APIError } from "../lib/api";

const signedIn = ref<boolean | null>(null);
const signInErr = ref<string | null>(null);

(async function () {
  if (window.location.pathname == "/signIn/redirect") {
    try {
      await doAPIRequest(
        "GET",
        "/signIn/redirect" + window.location.search,
        200
      );
    } catch (e) {
      signedIn.value = false;
      if (e instanceof APIError) {
        signInErr.value = e.message;
      } else if (e instanceof Error) {
        signInErr.value = e.message;
      } else {
        signInErr.value = String(e);
      }
    }
    window.history.replaceState({}, "", "/");
  }

  try {
    const user = (await doAPIRequest("GET", "/me", 200)) as User;
    currentUser.value = user;
    signedIn.value = true;
  } catch (e) {
    signedIn.value = false;
    if (e instanceof APIError) {
      signInErr.value = e.message;
    } else if (e instanceof Error) {
      signInErr.value = e.message;
    } else {
      signInErr.value = String(e);
    }
  }
})();
</script>

<template>
  <div v-if="signedIn === null">
    <h1>Signing you in, please wait...</h1>
  </div>
  <slot v-else-if="signedIn"></slot>
  <div v-else>
    <h1>Sign in failed</h1>
    <p>{{ signInErr }}</p>
  </div>
</template>

<style scoped></style>
