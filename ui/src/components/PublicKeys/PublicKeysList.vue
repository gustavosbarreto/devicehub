<template>
  <div>
    <DataTable
      v-bind="$attrs"
      :headers="headers"
      :items="publicKeys"
      :itemsPerPage="itemsPerPage"
      :nextPage="next"
      :previousPage="prev"
      :loading="loading"
      :totalCount="getNumberPublicKeys"
      :actualPage="page"
      :comboboxOptions="[10, 20, 50, 100]"
      @changeItemsPerPage="changeItemsPerPage"
      @clickNextPage="next"
      @clickPreviousPage="prev"
      data-test="publicKeys-list"
    >
      <template v-slot:rows>
        <tr v-for="(item, i) in publicKeys" :key="i" data-test="public-key-item">
          <td class="text-center" data-test="public-key-name">
            {{ item.name }}
          </td>

          <td class="text-center" data-test="public-key-fingerprint">
            {{ item.fingerprint }}
          </td>

          <td class="text-center" data-test="public-key-filter">
            <div v-if="isHostname(item.filter)">
              {{ formatHostnameFilter(item.filter) }}
            </div>
            <div v-else>
              <v-tooltip
                v-for="(tag, index) in item.filter.tags"
                :key="index"
                bottom
                :disabled="!showTag(tag)"
              >
                <template #activator="{ props }">
                  <v-chip
                    class="mr-1"
                    density="compact"
                    outlined
                    v-bind="props"
                  >
                    {{ displayOnlyTenCharacters(tag) }}
                  </v-chip>
                </template>

                <span v-if="showTag(tag)">
                  {{ tag }}
                </span>
              </v-tooltip>
            </div>
          </td>

          <td class="text-center" data-test="public-key-username">
            {{ formatUsername(item.username) }}
          </td>

          <td class="text-center" data-test="public-key-created-at">
            {{ formatDateFullAbrevied(item.created_at) }}
          </td>

          <td class="text-center" data-test="public-key-actions">
            <v-menu location="bottom" scrim eager>
              <template v-slot:activator="{ props }">
                <v-btn
                  v-bind="props"
                  variant="plain"
                  class="border rounded bg-v-theme-background"
                  density="comfortable"
                  size="default"
                  icon="mdi-format-list-bulleted"
                  data-test="public-key-actions"
                />
              </template>
              <v-list class="bg-v-theme-surface" lines="two" density="compact">
                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationFormDialogEdit"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <PublicKeyEdit
                        :keyObject="item"
                        :notHasAuthorization="!hasAuthorizationFormDialogEdit"
                        @update="refreshPublicKeys"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>

                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationFormDialogRemove"
                >
                  <template v-slot:activator="{ props }">
                    <div v-bind="props">
                      <PublicKeyDelete
                        :fingerprint="item.fingerprint"
                        :notHasAuthorization="!hasAuthorizationFormDialogRemove"
                        @update="refreshPublicKeys"
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
import { computed, ref, watch } from "vue";
import { actions, authorizer } from "../../authorizer";
import { filterType } from "../../interfaces/IFirewallRule";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import {
  displayOnlyTenCharacters,
  formatHostnameFilter,
  formatUsername,
} from "../../utils/string";
import { formatDateFullAbrevied } from "../../utils/formateDate";
import showTag from "../../utils/tag";
import DataTable from "../DataTable.vue";
import PublicKeyDelete from "./PublicKeyDelete.vue";
import PublicKeyEdit from "./PublicKeyEdit.vue";
import { INotificationsError } from "@/interfaces/INotifications";
import handleError from "@/utils/handleError";

const headers = [
  {
    text: "Name",
    value: "name",
  },
  {
    text: "Fingerprint",
    value: "fingerprint",
  },
  {
    text: "Filter",
    value: "filter",
  },
  {
    text: "Username",
    value: "username",
  },
  {
    text: "Created At",
    value: "created_at",
  },
  {
    text: "Actions",
    value: "actions",
  },
];
const store = useStore();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const publicKeys = computed(() => store.getters["publicKeys/list"]);
const getNumberPublicKeys = computed(
  () => store.getters["publicKeys/getNumberPublicKeys"],
);

const hasAuthorizationFormDialogEdit = computed(() => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(authorizer.role[role], actions.publicKey.edit);
  }
  return false;
});

const hasAuthorizationFormDialogRemove = computed(() => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.publicKey.remove,
    );
  }
  return false;
});

const getPublicKeysList = async (
  perPagaeValue: number,
  pageValue: number,
) => {
  if (store.getters["box/getStatus"]) {
    const data = {
      perPage: perPagaeValue,
      page: pageValue,
    };
    try {
      loading.value = true;
      const hasPublicKeys = await store.dispatch("publicKeys/fetch", data);

      if (!hasPublicKeys) {
        page.value--;
      }
      loading.value = false;
    } catch (error: unknown) {
      store.dispatch(
        "snackbar/showSnackbarErrorLoading",
        INotificationsError.publicKeyList,
      );
      handleError(error);
    }
  } else {
    store.dispatch("box/setStatus", false);
  }
};

const next = async () => {
  await getPublicKeysList(itemsPerPage.value, ++page.value);
};

const prev = async () => {
  try {
    if (page.value > 1) await getPublicKeysList(itemsPerPage.value, --page.value);
  } catch (error: unknown) {
    store.dispatch("snackbar/setSnackbarErrorDefault");
    handleError(error);
  }
};

const changeItemsPerPage = async (newItemsPerPage: number) => {
  itemsPerPage.value = newItemsPerPage;
};

watch(itemsPerPage, async () => {
  await getPublicKeysList(itemsPerPage.value, page.value);
});

const refreshPublicKeys = async () => {
  await store.dispatch("publicKeys/refresh");
};

const isHostname = (filter: filterType) => Object.prototype.hasOwnProperty.call(filter, "hostname");

defineExpose({ publicKeys, hasAuthorizationFormDialogEdit, hasAuthorizationFormDialogRemove, getPublicKeysList });
</script>
