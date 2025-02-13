import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import Sessions from "@/views/Sessions.vue";
import { namespacesApi, sessionsApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";

type SessionsWrapper = VueWrapper<InstanceType<typeof Sessions>>;

describe("Sessions View", () => {
  let wrapper: SessionsWrapper;

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

  const sessionObj = { data: [{
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
    position: { longitude: 0, latitude: 0 },
  }],
  headers: {
    "x-total-count": 1,
  } };

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");

    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockSessions = new MockAdapter(sessionsApi.getAxios());

    mockSessions.onGet("http://localhost:3000/api/sessions?page=1&per_page=10").reply(200, sessionObj);
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("namespaces/setNamespaces", res);
    store.commit("sessions/setSessions", sessionObj);

    wrapper = mount(Sessions, {
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

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="sessions-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sessions-list"]').exists()).toBe(true);
  });

  it("Renders the SessionList component", () => {
    expect(wrapper.findComponent({ name: "SessionList" }).exists()).toBe(true);
  });
});
