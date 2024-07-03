<template>
  <div>
    <DataTable
      :headers="headers"
      :items="sessions"
      :itemsPerPage="itemsPerPage"
      :nextPage="next"
      :previousPage="prev"
      :loading="loading"
      :totalCount="numberSessions"
      :actualPage="page"
      :comboboxOptions="[10, 20, 50, 100]"
      @changeItemsPerPage="changeItemsPerPage"
      @clickNextPage="next"
      @clickPreviousPage="prev"
      data-test="sessions-list"
    >
      <template v-slot:rows>
        <tr v-for="(session, index) in sessions" :key="index">
          <td class="text-center">
            <v-icon v-if="session.active" color="success">
              mdi-check-circle
            </v-icon>
            <v-tooltip location="bottom" v-else>
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props"> mdi-check-circle </v-icon>
              </template>
              <span>{{ lastSeen(session.last_seen) }}</span>
            </v-tooltip>
          </td>

          <td class="text-center" v-if="session.device">
            <p
              @click="redirectDevice(session.device.uid)"
              @keyup="redirectDevice(session.device.uid)"
              tabindex="0"
              class="link"
            >
              {{ session.device.name }}
            </p>
          </td>

          <td class="text-center">
            <v-tooltip location="bottom" v-if="!session.authenticated">
              <template v-slot:activator="{ props }">
                <span v-bind="props">{{ session.username }}</span>
              </template>
              <span v-if="!session.authenticated">Unauthorized</span>
            </v-tooltip>
            <span v-if="session.authenticated">{{ session.username }}</span>
          </td>

          <td class="text-center">
            <v-tooltip location="bottom" v-if="session.authenticated">
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props" color="success">
                  mdi-shield-check
                </v-icon>
              </template>
              <span>User has been authenticated</span>
            </v-tooltip>
            <v-tooltip bottom v-else>
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props" color="error"> mdi-shield-alert </v-icon>
              </template>
              <span>User has not been authenticated</span>
            </v-tooltip>
          </td>

          <td class="text-center">
            <v-code class="bg-tabs">
              {{ session.ip_address }}
            </v-code>
          </td>
          <td class="text-center">
            <span>{{ formatDateCompact(session.started_at) }}</span>
          </td>

          <td class="text-center">
            <span>{{ formatDateCompact(session.last_seen) }}</span>
          </td>

          <td class="text-center">
            <v-menu location="bottom" scrim eager>
              <template v-slot:activator="{ props }">
                <v-chip v-bind="props" density="comfortable" size="small">
                  <v-icon>mdi-dots-horizontal</v-icon>
                </v-chip>
              </template>
              <v-list class="bg-v-theme-surface" lines="two" density="compact">
                <v-list-item @click="redirectToSession(session.uid)">
                  <div class="d-flex align-center">
                    <div class="mr-2">
                      <v-icon> mdi-information </v-icon>
                    </div>

                    <v-list-item-title data-test="mdi-information-list-item">
                      Details
                    </v-list-item-title>
                  </div>
                </v-list-item>

                <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorizationPlay()">
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <SessionPlay
                        v-if="session.authenticated && session.recorded"
                        :is-active="session.active"
                        :uid="session.uid"
                        :device="session.device"
                        :notHasAuthorization="!hasAuthorizationPlay()"
                        :recorded="session.authenticated && session.recorded"
                        @update="refreshSessions"
                        data-test="sessionPlay-component"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>

                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationRemoveRecord()"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <SessionClose
                        v-if="session.active"
                        :uid="session.uid"
                        :device="session.device"
                        :notHasAuthorization="!hasAuthorizationRemoveRecord"
                        @update="refreshSessions"
                        data-test="sessionDeleteRecord-component"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>
              </v-list>
            </v-menu>
          </td>
        </tr>
      </template>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from "vue";
import axios, { AxiosError } from "axios";
import { useRouter } from "vue-router";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import { useStore } from "../../store";
import { formatDateCompact, lastSeen } from "../../utils/formateDate";
import DataTable from "../DataTable.vue";
import SessionClose from "./SessionClose.vue";
import SessionPlay from "./SessionPlay.vue";
import { INotificationsError } from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

const store = useStore();
const router = useRouter();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);

const headers = [
  {
    text: "Active",
    value: "active",
  },
  {
    text: "Device",
    value: "device",
  },
  {
    text: "Username",
    value: "username",
  },
  {
    text: "Authenticated",
    value: "authenticated",
  },
  {
    text: "IP Address",
    value: "ip_address",
  },
  {
    text: "Started",
    value: "started",
  },
  {
    text: "Last Seen",
    value: "last_seen",
  },
  {
    text: "Actions",
    value: "actions",
  },
];

const sessions = computed(() => store.getters["sessions/list"]);
const numberSessions = computed(
  () => store.getters["sessions/getNumberSessions"],
);

const getSessions = async (perPagaeValue: number, pageValue: number) => {
  if (!store.getters["box/getStatus"]) {
    try {
      loading.value = true;
      const hasSessions = await store.dispatch("sessions/fetch", {
        page: pageValue,
        perPage: perPagaeValue,
      });

      if (!hasSessions) {
        page.value--;
      }
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError;
        if (axiosError.response?.status === 403) {
          store.dispatch("snackbar/showSnackbarErrorAssociation");
        }
      } else {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.sessionList,
        );
        handleError(error);
      }
    } finally {
      loading.value = false;
    }
  } else {
    store.dispatch("box/setStatus", false);
  }
};

onMounted(async () => {
  await getSessions(itemsPerPage.value, page.value);
});

const next = async () => {
  await getSessions(itemsPerPage.value, ++page.value);
};

const prev = async () => {
  try {
    if (page.value > 1) await getSessions(itemsPerPage.value, --page.value);
  } catch (error) {
    store.dispatch("snackbar/setSnackbarErrorDefault");
  }
};

const changeItemsPerPage = async (newItemsPerPage: number) => {
  itemsPerPage.value = newItemsPerPage;
};

watch(itemsPerPage, async () => {
  await getSessions(itemsPerPage.value, page.value);
});

const redirectToSession = (sessionUid: string) => {
  router.push({ name: "detailsSession", params: { id: sessionUid } });
};

const redirectDevice = (deviceUid: string) => {
  router.push({ name: "detailsDevice", params: { id: deviceUid } });
};

const refreshSessions = async () => {
  await getSessions(itemsPerPage.value, page.value);
};

const hasAuthorizationRemoveRecord = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.session.removeRecord,
    );
  }

  return false;
};

const hasAuthorizationPlay = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(authorizer.role[role], actions.session.play);
  }

  return false;
};
</script>

<style scoped>
.hover-text {
  cursor: pointer;
  animation: fadeIn 0.5s;
}

.hover-text:hover,
.hover-text:focus {
  text-decoration: underline;
}

.link {
  text-decoration: underline;
  cursor: pointer;
}
</style>
