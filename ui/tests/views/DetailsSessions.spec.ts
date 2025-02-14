import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createRouter, createWebHistory } from "vue-router";
import DetailsSessions from "@/views/DetailsSessions.vue";
import { namespacesApi, usersApi, sessionsApi } from "@/api/http";
import { store, key } from "@/store";
import { routes } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type DetailsSessionsWrapper = VueWrapper<InstanceType<typeof DetailsSessions>>;

describe("Details Sessions", () => {
  let wrapper: DetailsSessionsWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  let mockSessions: MockAdapter;

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

  const sessionObj = {
    uid: "1",
    device_uid: "1",
    device: {
      uid: "1",
      name: "00-00-00-00-00-01",
      identity: {
        mac: "00-00-00-00-00-01",
      },
      info: {
        id: "manjaro",
        pretty_name: "Manjaro Linux",
        version: "latest",
        arch: "amd64",
        platform: "docker",
      },
      public_key: "",
      tenant_id: "fake-tenant-data",
      last_seen: "0",
      online: true,
      namespace: "dev",
      status: "accepted",
      status_updated_at: "0",
      created_at: "0",
      remote_addr: "192.168.0.1",
      position: { latitude: 0, longitude: 0 },
      tags: [],
      public_url: false,
      public_url_address: "",
      acceptable: false,
    },
    tenant_id: "fake-tenant-data",
    username: "test",
    ip_address: "192.168.0.1",
    started_at: "",
    last_seen: "",
    active: false,
    authenticated: true,
    recorded: true,
    type: "none",
    term: "none",
    position: { longitude: 0, latitude: 0 } };

  let router;

  beforeEach(async () => {
    router = createRouter({
      history: createWebHistory(),
      routes,
    });

    router.push("/sessions/1");

    await router.isReady();

    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");

    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockSessions = new MockAdapter(sessionsApi.getAxios());

    mockSessions.onGet("http://localhost:3000/api/sessions/1").reply(200, sessionObj);
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("namespaces/setNamespaces", res);
    store.commit("sessions/setSession", sessionObj);

    wrapper = mount(DetailsSessions, {
      global: {
        plugins: [[store, key], vuetify, [router], SnackbarPlugin],
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

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="sessionUid-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sessionUser-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sessionAuthenticated-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sessionIpAddress-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sessionStartedAt-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sessionLastSeen-field"]').exists()).toBe(true);
  });
});
