import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import Devices from "@/views/Devices.vue";
import { namespacesApi, usersApi, devicesApi } from "@/api/http";
import { store, key } from "@/store";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";

type DevicesWrapper = VueWrapper<InstanceType<typeof Devices>>;

describe("Devices View", () => {
  let wrapper: DevicesWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  let mockDevices: MockAdapter;

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];
  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    settings: {
      session_record: true,
    },
    max_devices: 3,
    devices_count: 3,
    created_at: "",
  };

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  const session = true;

  const res = {
    data: [namespaceData],
    headers: {
      "x-total-count": 1,
    },
  };

  const stats = {
    data: {
      registered_devices: 2,
      online_devices: 1,
      active_sessions: 0,
      pending_devices: 24,
      rejected_devices: 1,
    },
  };

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");

    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());

    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("namespaces/setNamespaces", res);
    store.commit("stats/setStats", stats);

    wrapper = mount(Devices, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    wrapper.unmount();
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

  it("Renders the template with data", async () => {
    expect(wrapper.find('[data-test="search-text"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="device-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-header-component-group"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-table-component"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="boxMessageDevice-component"]').exists()).toBe(true);
  });
});
