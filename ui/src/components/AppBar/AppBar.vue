<template>
  <PaywallChat v-model="chatSupportPaywall" />
  <v-app-bar
    flat
    floating
    class="bg-background border-b-thin"
    data-test="app-bar"
  >
    <v-app-bar-nav-icon
      class="hidden-lg-and-up"
      @click.stop="showNavigationDrawer = !showNavigationDrawer"
      aria-label="Toggle Menu"
      data-test="menu-toggle"
    />
    <v-icon icon="mdi-server-network" class="ml-4 hidden-md-and-down" />

    <v-breadcrumbs :items="breadcrumbItems" class="hidden-md-and-down" data-test="breadcrumbs">
      <template v-slot:divider>
        <v-icon icon="mdi-chevron-right" />
      </template>
    </v-breadcrumbs>

    <v-spacer />

    <v-tooltip
      location="bottom"
      class="text-center"
    >
      <template v-slot:activator="{ props }">
        <v-btn
          v-bind="props"
          size="medium"
          color="primary"
          aria-label="community-help-icon"
          icon="mdi-help-circle"
          @click="openShellhubHelp()"
          data-test="support-btn"
        />
      </template>
      <span>Need assistance? Click here for support.</span>
    </v-tooltip>

    <Notification data-test="notification-component" />

    <v-menu>
      <template v-slot:activator="{ props }">
        <v-btn
          color="primary"
          v-bind="props"
          append-icon="mdi-menu-down"
          class="pl-2 pr-2 mr-4"
          data-test="user-menu-btn"
        >
          <UserIcon size="1.5rem" :email="userEmail" data-test="user-icon" />
        </v-btn>
      </template>
      <v-list class="bg-v-theme-surface">
        <v-list-item
          v-for="item in menu"
          :key="item.title"
          :value="item"
          :data-test="item.title"
          @click="triggerClick(item)"
        >
          <div class="d-flex align-center">
            <v-icon
              :icon="item.icon"
              class="mr-2"
            />

            <v-list-item-title>
              {{ item.title }}
            </v-list-item-title>
          </div>
        </v-list-item>

        <v-divider />

        <v-list-item density="compact">
          <v-switch
            label="Dark Mode"
            :model-value="isDarkMode"
            :onchange="toggleDarkMode"
            data-test="dark-mode-switch"
            density="comfortable"
            color="primary"
            inset
            hide-details
          />
        </v-list-item>
      </v-list>
    </v-menu>
  </v-app-bar>
</template>

<script setup lang="ts">
import {
  computed,
  ref,
} from "vue";
import { useRouter, useRoute, RouteLocationRaw, RouteLocation } from "vue-router";
import { useChatWoot } from "@productdevbook/chatwoot/vue";
import { useEventListener } from "@vueuse/core";
import { useStore } from "../../store";
import { createNewClient } from "../../api/http";
import handleError from "../../utils/handleError";
import UserIcon from "../User/UserIcon.vue";
import Notification from "./Notifications/Notification.vue";
import PaywallChat from "../User/PaywallChat.vue";
import { envVariables } from "@/envVariables";

type MenuItem = {
  title: string;
  icon: string;
  type: string;
  path: RouteLocationRaw;
  method: () => void;
};

type BreadcrumbItem = {
  title: string;
  href: string;
};
const store = useStore();
const router = useRouter();
const route = useRoute();
const getStatusDarkMode = computed(
  () => store.getters["layout/getStatusDarkMode"],
);
const tenant = computed(() => store.getters["auth/tenant"]);
const userEmail = computed(() => store.getters["auth/email"]);
const userId = computed(() => store.getters["auth/id"]);
const currentUser = computed(() => store.getters["auth/currentUser"]);
const billingActive = computed(() => store.getters["billing/active"]);
const identifier = computed(() => store.getters["support/getIdentifier"]);
const isDarkMode = ref(getStatusDarkMode.value === "dark");
const chatSupportPaywall = ref(false);
const showNavigationDrawer = defineModel<boolean>();

const triggerClick = (item: MenuItem): void => {
  switch (item.type) {
    case "path":
      router.push(item.path);
      break;
    case "method":
      item.method();
      break;
    default:
      break;
  }
};

const logout = async () => {
  try {
    await store.dispatch("auth/logout");
    await store.dispatch("stats/clear");
    await store.dispatch("namespaces/clearNamespaceList");
    await router.push({ name: "Login" });
    createNewClient();
  } catch (error: unknown) {
    handleError(error);
  }
};

const toggleDarkMode = () => {
  isDarkMode.value = !isDarkMode.value;
  store.dispatch("layout/setStatusDarkMode", isDarkMode.value);
};

const openChatwoot = async (): Promise<void> => {
  const { setUser, setConversationCustomAttributes, toggle } = useChatWoot();

  await store.dispatch("support/get", tenant.value);

  setUser(userId.value, {
    name: currentUser.value,
    email: userEmail.value,
    identifier_hash: identifier.value,
  });

  useEventListener(window, "chatwoot:on-message", () => {
    setConversationCustomAttributes({
      namespace: store.getters["namespaces/get"].name,
      tenant: tenant.value,
      domain: window.location.hostname,
    });
  });

  store.commit("support/setCreatedStatus", true);
  toggle("open");
};

const openPaywall = (): void => {
  chatSupportPaywall.value = true;
};

const redirectToGitHub = (): void => {
  window.open("https://github.com/shellhub-io/shellhub/issues/new/choose", "_blank");
};

const openShellhubHelp = async (): Promise<void> => {
  switch (true) {
    case envVariables.isCloud && billingActive.value:
      await openChatwoot();
      break;

    case envVariables.isCommunity || (envVariables.isCloud && !billingActive.value):
      openPaywall();
      break;

    case envVariables.isEnterprise:
      redirectToGitHub();
      break;

    default:
      throw new Error("Unsupported environment configuration.");
  }
};

const menu = [
  {
    title: "Settings",
    type: "path",
    path: "/Settings",
    icon: "mdi-cog",
    // eslint-disable-next-line no-void
    method: () => void 0,
  },
  {
    title: "Logout",
    type: "method",
    icon: "mdi-logout",
    path: "",
    method: logout,
  },
];

const generateBreadcrumbs = (route: RouteLocation): BreadcrumbItem[] => {
  const breadcrumbs: BreadcrumbItem[] = [];
  route.matched.forEach((match) => {
    if (match.name) {
      const title = (match.name as string).replace(/([a-z])([A-Z])/g, "$1 $2");
      breadcrumbs.push({
        title,
        href: match.path,
      });
    }
  });
  return breadcrumbs;
};

const breadcrumbItems = computed(() => generateBreadcrumbs(route));

defineExpose({ openShellhubHelp, chatSupportPaywall, logout, isDarkMode, breadcrumbItems, currentUser, identifier });
</script>
