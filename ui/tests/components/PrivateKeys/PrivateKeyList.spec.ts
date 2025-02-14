import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PrivateKeyList from "@/components/PrivateKeys/PrivateKeyList.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type PrivateKeyListWrapper = VueWrapper<InstanceType<typeof PrivateKeyList>>;

describe("Private Key List", () => {
  let wrapper: PrivateKeyListWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

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
      connection_announcement: "",
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

  const PrivateKeys = [
    {
      name: "",
      data: "",
      id: 1,
    },
    {
      name: "",
      data: "",
      id: 2,
    },
    {
      name: "",
      data: "",
      id: 3,
    },
  ];

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    wrapper = mount(PrivateKeyList, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("security/setSecurity", session);
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

  it("Renders components", async () => {
    expect(wrapper.find('[data-test="no-private-key-warning"]').exists()).toBe(true);
    store.commit("privateKey/fetchPrivateKey", PrivateKeys);
    await flushPromises();
    expect(wrapper.find('[data-test="privateKey-thead"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privateKey-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privateKey-fingerprint"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privateKey-actions"]').exists()).toBe(true);
  });
});
