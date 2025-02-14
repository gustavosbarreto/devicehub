import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import { store, key } from "@/store";
import PaywallChat from "@/components/User/PaywallChat.vue";
import { router } from "@/router";
import { namespacesApi, devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

const vuetify = createVuetify();

const namespaceData = {
  name: "test",
  owner: "xxxxxxxx",
  tenant_id: "fake-tenant-data",
  members: [{ id: "xxxxxxxx", username: "test", role: "owner" }],
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

const stats = {
  registered_devices: 3,
  online_devices: 1,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

describe("PaywallChat", () => {
  let wrapper: VueWrapper<InstanceType<typeof PaywallChat>>;
  let mockNamespace: MockAdapter;
  let mockDevices: MockAdapter;

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant-data");

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());
    mockNamespace.onGet("/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockDevices.onGet("/api/stats").reply(200, stats);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(PaywallChat, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      attachTo: document.body,
    });
  });

  it("is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("defines data properties", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("renders dialog elements when opened", async () => {
    wrapper.vm.dialog = true;
    await flushPromises();

    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="card-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="icon-chat"]').exists()).toBe(true);
    expect(dialog.find('[data-test="upgrade-heading"]').exists()).toBe(true);
    expect(dialog.find('[data-test="upgrade-description"]').exists()).toBe(true);
    expect(dialog.find('[data-test="upgrade-description2"]').exists()).toBe(true);
    expect(dialog.find('[data-test="link-anchor"]').exists()).toBe(true);
    expect(dialog.find('[data-test="card-actions"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="upgrade-btn"]').exists()).toBe(true);
  });

  it("ensures the upgrade button has correct href", () => {
    wrapper.vm.dialog = true;
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="upgrade-btn"]').attributes("href")).toBe("www.shellhub.io/pricing");
  });
});
