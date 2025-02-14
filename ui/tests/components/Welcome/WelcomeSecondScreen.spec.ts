import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import WelcomeSecondScreen from "@/components/Welcome/WelcomeSecondScreen.vue";
import { envVariables } from "@/envVariables";
import { router } from "@/router";
import { namespacesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { INotificationsCopy } from "@/interfaces/INotifications";

type WelcomeSecondScreenWrapper = VueWrapper<InstanceType<typeof WelcomeSecondScreen>>;

describe("Welcome Second Screen", () => {
  let wrapper: WelcomeSecondScreenWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
    name: "user",
    owner: "xxxxxxxx",
    tenant_id: "fake-tenant-data",
    members,
    max_devices: 3,
    devices_count: 3,
    devices: 2,
    created_at: "",
  };

  const authData = {
    status: "",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
    role: "owner",
  };

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(WelcomeSecondScreen, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Renders the components", async () => {
    expect(wrapper.find('[data-test="welcome-second-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-second-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-second-run-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="command-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-second-link-docs"]').exists()).toBe(true);
  });

  it("Should copy the command to clipboard when copyCommand is called", async () => {
    const storeSpy = vi.spyOn(store, "dispatch");
    wrapper.vm.copyCommand();
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith("snackbar/showSnackbarCopy", INotificationsCopy.tenantId);
  });
});
