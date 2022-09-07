<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref} from "vue";
import {CompletedChallenges, Scoreboard, Team} from "../lib/types";
import {doAPIRequest} from "../lib/api";
import {Dataset, datasets} from "../lib/datasetState";

(async function () {
  datasets.value = (await doAPIRequest("GET", "/datasets", 200)) as Dataset[];
})();

const teams = ref<Team[] | null>(null);
const scoreboard = ref<Scoreboard | null>(null);
const completedChallenges = ref<CompletedChallenges | null>(null);
const error = ref<string | null>(null);

const teamNames = computed(() => {
  if (teams.value === null) {
    return {};
  }
  let result: Record<string, string> = {};
  for (const team of teams.value) {
    result[team.id] = team.name;
  }
  return result;
});
const completeByTeam = computed(() => {
  if (teams.value === null || completedChallenges.value === null) {
    return {};
  }
  let result: Record<string, number> = {};
  for (const team of teams.value) {
    result[team.id] = 0;
  }
  for (const ds of Object.keys(completedChallenges.value)) {
    for (const q of Object.keys(completedChallenges.value[ds])) {
      for (const team of Object.keys(completedChallenges.value[ds][q])) {
        if (completedChallenges.value[ds][q][team]) {
          result[team]++;
        }
      }
    }
  }
  return result;
});
const incomplete = computed(() => {
  console.log(completedChallenges.value, datasets.value);
  if (completedChallenges.value === null || datasets.value === null) {
    return {};
  }
  console.log(completedChallenges.value);
  let result: Record<string, Record<string, string[]>> = {};
  for (const ds of Object.keys(completedChallenges.value)) {
    const dsName = datasets.value!.find(x => x.id === ds)!.name;
    result[dsName] = {};
    // @ts-expect-error go home typescript, you're drunk
    for (const qId of Object.keys(completedChallenges.value[ds])) {
      // @ts-expect-error
      const qName = datasets.value!.find(x => x.id === ds)!.queries.find(x => x.id === qId)!.name;
      result[dsName][qName] = [];
      for (const team of Object.keys(completedChallenges.value[ds][qId])) {
        if (!completedChallenges.value[ds][qId][team]) {
          result[dsName][qName].push(teamNames.value[team]);
        }
      }
    }
  }
  return result;
});

let updateInterval: number;
onMounted(() => {
  async function update() {
    try {
      const [td, sd, ccd] = await Promise.all([
        doAPIRequest<Team[]>("GET", "/teams", 200),
        doAPIRequest<Scoreboard>("GET", "/scoreboard", 200),
        doAPIRequest<CompletedChallenges>("GET", "/completedChallenges", 200)
      ]);
      teams.value = td;
      scoreboard.value = sd;
      completedChallenges.value = ccd;
      error.value = null;
    } catch (e) {
      error.value = String(e);
    } finally {
      updateInterval = setTimeout(update, 15_000);
    }
  }
  update();
});
onUnmounted(() => {
  clearTimeout(updateInterval);
});

const page = ref(0);
let paused = false;
const MAX_PAGE = 2;
let pageInterval: number;
function flip() {
  if (page.value === MAX_PAGE) {
    page.value = 0;
  } else {
    page.value++;
  }
}

function onKey(e: KeyboardEvent) {
  switch (e.code) {
    case "ArrowRight":
      flip();
      break;
    case "ArrowLeft":
      if (page.value === 0) {
        page.value = MAX_PAGE
      } else {
        page.value--
      }
      break;
    case "KeyP":
      if (paused) {
        pageInterval = setInterval(flip, 15_000);
      } else {
        clearInterval(pageInterval);
      }
      break;
  }
}

onMounted(() => {
  pageInterval = setInterval(flip, 15_000);
  window.addEventListener("keydown", onKey);
})
onUnmounted(() =>{
  clearInterval(pageInterval);
  window.removeEventListener("keydown", onKey);
})
</script>

<template>
  <div class="wrapper">
    <Transition>
      <div v-if="page === 0" class="slide" id="1">
        <h1>SCORES</h1>
        <table>
          <tr v-for="(points, teamId) in scoreboard">
            <td>{{ teamNames[teamId] }}</td>
            <td>{{ points }}</td>
          </tr>
        </table>
      </div>
      <div v-else-if="page === 1" class="slide" id="2">
        <h1>COMPLETED CHALLENGES</h1>
       <table>
         <tr v-for="(qs, teamId) in completeByTeam">
           <td>{{ teamNames[teamId] }}</td>
           <td>{{ qs }}</td>
         </tr>
       </table>
      </div>
      <div v-else-if="page === 2" class="slide" id="3">
        <h1>REMAINING CHALLENGES</h1>
        <div>
          <div v-for="(queries, dsName) in incomplete">
            <h2>{{ dsName }}</h2>
            <table>
              <tr v-for="(teams, qName) in queries">
                <td v-if="teams.length > 0">{{ qName }}</td>
                <td>{{ teams.join(", ") }}</td>
              </tr>
            </table>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.wrapper {
  overflow-x: hidden;
}
.slide {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%,-50%);
  font-size: 1.4rem;
  line-height: 1.9em;
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 50vw;
}
table {
  font-size: 2rem;
  margin: 0 !important;
  width: 100%;
}
td:nth-child(1) {
  text-align: left;
}
td:nth-child(2) {
  font-weight: bold;
  text-align: right;
  padding-left: 0.3rem;
}
ul {
  padding: 0;
}
.v-enter-from {
  /*transform: translateX(100vw);*/
  opacity: 0;
}
.v-enter-active, .v-leave-active {
  transition: transform 1s ease-out, opacity 1s;
}
.v-leave-to {
  /*transform: translateX(-100vw);*/
  opacity: 0;
}
</style>