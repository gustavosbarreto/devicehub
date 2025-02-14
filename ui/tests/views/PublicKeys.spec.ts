import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PublicKeys from "@/views/PublicKeys.vue";
import { namespacesApi, usersApi, sshApi } from "@/api/http";
import { store, key } from "@/store";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";

type PublicKeysWrapper = VueWrapper<InstanceType<typeof PublicKeys>>;

describe("Public Keys", () => {
  let wrapper: PublicKeysWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  let mockSsh: MockAdapter;

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

  const publicKeys = {
    data: [
      {
        data: "",
        fingerprint: "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:01",
        created_at: "",
        tenant_id: "00000000-0000-4000-0000-000000000000",
        name: "public-key-test",
        username: ".*",
        filter: {
          hostname: ".*",
        },
      },
    ],
    headers: {
      "x-total-count": 1,
    },
  };

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");

    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockSsh = new MockAdapter(sshApi.getAxios());

    mockSsh.onGet("http://localhost:3000/api/sshkeys/public-keys?filter=&page=1&per_page=10").reply(200, publicKeys);
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("namespaces/setNamespaces", res);

    wrapper = mount(PublicKeys, {
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
    expect(wrapper.find('[data-test="public-keys-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-keys-components"]').exists()).toBe(true);
  });

  it("Renders the PublicKeyAdd component", () => {
    expect(wrapper.findComponent({ name: "PublicKeyAdd" }).exists()).toBe(true);
  });

  it("Refreshes the public keys", async () => {
    const refreshSpy = vi.spyOn(store, "dispatch");
    await wrapper.vm.refresh();
    expect(refreshSpy).toHaveBeenCalledWith("publicKeys/refresh");
  });
});
