import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import TeamApiKeys from "@/views/TeamApiKeys.vue";
import { namespacesApi, usersApi, apiKeysApi, devicesApi } from "@/api/http";
import { store, key } from "@/store";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";

type TeamApiKeysWrapper = VueWrapper<InstanceType<typeof TeamApiKeys>>;

describe("Team Api Keys", () => {
  let wrapper: TeamApiKeysWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  let mockDevices: MockAdapter;

  let mockApiKeys: MockAdapter;

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

  const apiKeys = [
    {
      name: "fake-api-key",
      tenant_id: "00000000-0000-4000-0000-000000000000",
      role: "owner",
      created_by: "xxxxxxxx",
      created_at: "",
      updated_at: "",
      expires_in: "",
    },
  ];

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");

    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockApiKeys = new MockAdapter(apiKeysApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());

    mockApiKeys.onGet("http://localhost:3000/api/namespaces/api-key?page=1&per_page=10").reply(200, apiKeys);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("namespaces/setNamespaces", res);
    store.commit("stats/setStats", stats);

    wrapper = mount(TeamApiKeys, {
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
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="api-key-generate"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="api-key-list"]').exists()).toBe(true);
  });
});
