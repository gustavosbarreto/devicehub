import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import TeamMembers from "@/views/TeamMembers.vue";
import { namespacesApi, usersApi, apiKeysApi } from "@/api/http";
import { store, key } from "@/store";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";

type TeamMembersWrapper = VueWrapper<InstanceType<typeof TeamMembers>>;

describe("Team Members", () => {
  let wrapper: TeamMembersWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  let mockApiKeys: MockAdapter;

  const members = [
    {
      id: "507f1f77bcf86cd799439011",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = { data: {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant",
    members,
    settings: {
      session_record: true,
      connection_announcement: "",
    },
    max_devices: 3,
    devices_count: 3,
    created_at: "",
  },
  };

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant",
    email: "test@test.com",
    id: "507f1f77bcf86cd799439011",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  const session = true;

  const getKeyResponse = [
    {
      id: "3e5a5194-9dec-4a32-98db-7434c6d49df1",
      tenant_id: "fake-tenant",
      user_id: "507f1f77bcf86cd799439011",
      name: "my api key",
      expires_in: 1707958989,
    },
    {
      id: "3e5a5194-9dec-4a32-98db-7434c6d49df2",
      tenant_id: "fake-tenant",
      user_id: "507f1f77bcf86cd799439011",
      name: "my api key",
      expires_in: 1707958989,
    },
  ];

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockApiKeys = new MockAdapter(apiKeysApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockApiKeys.onGet("http://localhost:3000/api/namespaces/fake-tenant/api-key").reply(200, getKeyResponse, { "x-total-count": 2 });

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("security/setSecurity", session);
    store.commit("auth/setKeyList", { data: getKeyResponse, headers: { "x-total-count": 2 } });

    wrapper = mount(TeamMembers, {
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
    expect(wrapper.find('[data-test="member-invite"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="member-list"]').exists()).toBe(true);
  });
});
