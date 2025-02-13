import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import QuickConnection from "@/components/QuickConnection/QuickConnection.vue";
import { envVariables } from "@/envVariables";
import { router } from "@/router";
import { namespacesApi, devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type QuickConnectionWrapper = VueWrapper<InstanceType<typeof QuickConnection>>;

describe("Quick Connection", () => {
  const node = document.createElement("div");
  node.setAttribute("id", "app");
  document.body.appendChild(node);

  let wrapper: QuickConnectionWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockDevices: MockAdapter;

  const devices = [
    {
      uid: "a582b47a42e",
      name: "39-5e-2b",
      identity: {
        mac: "00:00:00:00:00:00",
      },
      info: {
        id: "linuxmint",
        pretty_name: "Linux Mint 19.3",
        version: "",
      },
      public_key: "----- PUBLIC KEY -----",
      tenant_id: "fake-tenant-data",
      last_seen: "2020-05-20T19:58:53.276Z",
      online: true,
      namespace: "user",
      status: "accepted",
      tags: ["test-tag"],
    },
  ];

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

  const stats = {
    registered_devices: 2,
    online_devices: 1,
    active_sessions: 0,
    pending_devices: 0,
    rejected_devices: 0,
  };

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockDevices
      // eslint-disable-next-line vue/max-len
      .onGet("http://localhost:3000/api/devices?filter=W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJvbmxpbmUiLCJvcGVyYXRvciI6ImVxIiwidmFsdWUiOnRydWV9fV0%3D&per_page=10&status=accepted")
      .reply(200, devices);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(QuickConnection, {
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

  it("Renders the dialog open button and other key elements", async () => {
    const dialog = new DOMWrapper(document.body);

    expect(wrapper.find('[data-test="quick-connection-open-btn"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="quick-connection-open-btn"]').trigger("click");

    expect(dialog.find('[data-test="search-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="hostname-header"]').exists()).toBe(true);
    expect(dialog.find('[data-test="os-header"]').exists()).toBe(true);
    expect(dialog.find('[data-test="sshid-header"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tags-header"]').exists()).toBe(true);
    expect(dialog.find('[data-test="connect-icon"]').exists()).toBe(true);
    expect(dialog.find('[data-test="copy-sshid-instructions"]').exists()).toBe(true);
    expect(dialog.find('[data-test="connect-icon"]').exists()).toBe(true);
    expect(dialog.find('[data-test="navigate-up-icon"]').exists()).toBe(true);
    expect(dialog.find('[data-test="navigate-down-icon"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("keyboardMacros function toggles dialog value on Ctrl + K keydown", async () => {
    const event = new KeyboardEvent("keydown", { ctrlKey: true, key: "k" });

    dispatchEvent(event);

    expect(wrapper.find('[data-test="quick-connection-open-btn"]').exists()).toBe(true);
  });

  it("Checks if the fetch function handles error on failure", async () => {
    mockDevices.reset();

    mockDevices
    // eslint-disable-next-line vue/max-len
      .onGet("http://localhost:3000/api/devices?filter=W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJvbmxpbmUiLCJvcGVyYXRvciI6ImVxIiwidmFsdWUiOnRydWV9fV0%3D&per_page=10&status=accepted")
      .reply(403);

    const storeSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="quick-connection-open-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith("snackbar/showSnackbarErrorDefault");
  });
});
