import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PublicKeysList from "@/components/PublicKeys/PublicKeysList.vue";
import { namespacesApi, usersApi, sshApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type PublicKeysListWrapper = VueWrapper<InstanceType<typeof PublicKeysList>>;

describe("Public Key List", () => {
  let wrapper: PublicKeysListWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  let mockSsh: MockAdapter;

  const members = [
    {
      id: "507f1f77bcf86cd799439011",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
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

  const res = {
    data: [
      {
        data: `LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlHZk1BMEdDU3FHU0liM0RRR
        UJBUVVBQTRHTkFEQ0JpUUtCZ1FDTHlMRjdUT0VhbFphdE5RcUlRQTdtd1d3WQoxcU5wN
        Elzcm8ya04zdTdIUEh1WmN1VGxlRS94MWpXNFZSTUJTaWMrZzNxak5VaS9YVCtMYndzO
        VRmN0ZqVUxLCnEwUzduRGRWNGRjYmYwdHJROWM4K3gvOHEvcURUWk13SC9sRXdVQ1hXQ
        TF5YW5JVjllT1RMUURld0VnRUFHYzQKeDFOQ2NaWTVUY3dZanprNTN3SURBUUFCCi0t
        LS0tRU5EIFBVQkxJQyBLRVktLS0tLQ==`,
        fingerprint: "fake-fingerprint",
        created_at: "2020-05-01T00:00:00.000Z",
        tenant_id: "fake-tenant",
        name: "example",
        filter: {
          hostname: ".*",
        },
        username: ".*",
      },
    ],
    headers: {
      "x-total-count": 1,
    },
  };

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockSsh = new MockAdapter(sshApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockSsh.onGet("http://localhost:3000/api/sshkeys/public-keys?filter=&page=1&per_page=10").reply(200, res);

    store.commit("auth/authSuccess", authData);
    store.commit("publicKeys/setPublicKeys", res);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("security/setSecurity", session);
    wrapper = mount(PublicKeysList, {
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

  it("Renders components", async () => {
    expect(wrapper.find('[data-test="publicKeys-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-item"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-fingerprint"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-filter"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-username"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-created-at"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="public-key-actions"]').exists()).toBe(true);
  });

  it("Handles authorization for editing and removing keys", async () => {
    // Mock the store to have different roles for the tests
    const roleOwner = "owner";
    const roleUser = "user";
    store.state.auth.role = roleOwner;
    expect(wrapper.vm.hasAuthorizationFormDialogEdit).toBeTruthy();
    expect(wrapper.vm.hasAuthorizationFormDialogRemove).toBeTruthy();

    store.state.auth.role = roleUser;
    expect(wrapper.vm.hasAuthorizationFormDialogEdit).toBeFalsy();
    expect(wrapper.vm.hasAuthorizationFormDialogRemove).toBeFalsy();
  });

  it("Checks if the public key list is not empty", () => {
    expect(wrapper.vm.publicKeys.length).toBeGreaterThan(0);
  });

  it("Checks if the public key has correct properties", () => {
    const publicKey = wrapper.vm.publicKeys[0];
    expect(publicKey).toHaveProperty("data");
    expect(publicKey).toHaveProperty("fingerprint");
    expect(publicKey).toHaveProperty("created_at");
    expect(publicKey).toHaveProperty("tenant_id");
    expect(publicKey).toHaveProperty("name");
    expect(publicKey).toHaveProperty("filter");
    expect(publicKey).toHaveProperty("username");
  });

  it("Checks if the public key filter is a hostname", () => {
    const publicKey = wrapper.vm.publicKeys[0];
    expect(publicKey.filter).toHaveProperty("hostname");
  });
});
