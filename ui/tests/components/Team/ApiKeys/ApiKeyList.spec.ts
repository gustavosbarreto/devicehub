import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import moment from "moment";
import ApiKeyList from "@/components/Team/ApiKeys/ApiKeyList.vue";
import { namespacesApi, usersApi, apiKeysApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type ApiKeyListWrapper = VueWrapper<InstanceType<typeof ApiKeyList>>;

describe("Api Key List", () => {
  let wrapper: ApiKeyListWrapper;

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

  const getKeyResponse = [
    {
      name: "aaaa2",
      tenant_id: "00000-0000-0000-0000-00000000000",
      role: "administrator",
      created_by: "66562f80daba745a106393b5",
      created_at: "2024-06-07T12:10:56.531Z",
      updated_at: "2024-06-07T12:31:03.505Z",
      expires_in: 1720354256,
    },
    {
      name: "aaaa2",
      tenant_id: "00000-0000-0000-0000-00000000000",
      role: "administrator",
      created_by: "66562f80daba745a106393b5",
      created_at: "2024-06-07T12:10:56.531Z",
      updated_at: "2024-06-07T12:31:03.505Z",
      expires_in: 1720354256,
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
    mockApiKeys.onGet("http://localhost:3000/api/namespaces/api-key?page=1&per_page=10").reply(200, getKeyResponse, { "x-total-count": 2 });

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("security/setSecurity", session);
    store.commit("apiKeys/setKeyList", { data: getKeyResponse, headers: { "x-total-count": 2 } });

    wrapper = mount(ApiKeyList, {
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
    expect(wrapper.find('[data-test="api-key-list"]').exists()).toBe(true);
  });

  describe("expiration formatting", () => {
    it("Returns 'Never' when unixTime is -1", () => {
      const result = wrapper.vm.formatDate(-1);
      expect(result).toBe("Never");
    });

    it("Formats unixTime into 'Expires on <date>' when unixTime is in the future", () => {
      const futureUnixTime = moment().add(1, "day").unix();
      const result = wrapper.vm.formatDate(futureUnixTime);
      const expected = `Expires on ${moment.unix(futureUnixTime).format("MMM D YYYY")}.`;
      expect(result).toBe(expected);
    });

    it("Formats unixTime into 'Expired on <date>' when unixTime is in the past", () => {
      const pastUnixTime = moment().subtract(1, "day").unix();
      const result = wrapper.vm.formatDate(pastUnixTime);
      const expected = `Expired on ${moment.unix(pastUnixTime).format("MMM D YYYY")}.`;
      expect(result).toBe(expected);
    });

    it("Returns false when unixTime is -1", () => {
      const result = wrapper.vm.formatKey(-1);
      expect(result).toBe(false);
    });

    it("Returns true when unixTime is in the past", () => {
      const pastUnixTime = moment().subtract(1, "day").unix();
      const result = wrapper.vm.formatKey(pastUnixTime);
      expect(result).toBe(true);
    });

    it("Returns false when unixTime is in the future", () => {
      const futureUnixTime = moment().add(1, "day").unix();
      const result = wrapper.vm.formatKey(futureUnixTime);
      expect(result).toBe(false);
    });

    describe("expired vs will expire items formatting", () => {
      it("Formats items that will expire correctly", () => {
        const futureUnixTime = moment().add(1, "day").unix();
        const item = { expires_in: futureUnixTime };

        const formatted = wrapper.vm.formatDate(item.expires_in);
        const expected = `Expires on ${moment.unix(futureUnixTime).format("MMM D YYYY")}.`;

        expect(formatted).toBe(expected);
      });

      it("Formats expired items correctly", () => {
        const pastUnixTime = moment().subtract(1, "day").unix();
        const item = { expires_in: pastUnixTime };

        const formatted = wrapper.vm.formatDate(item.expires_in);
        const expected = `Expired on ${moment.unix(pastUnixTime).format("MMM D YYYY")}.`;

        expect(formatted).toBe(expected);
      });

      it("Formats items that will not expire correctly", () => {
        const item = { expires_in: -1 };

        const formatted = wrapper.vm.formatDate(item.expires_in);
        const expected = "Never";

        expect(formatted).toBe(expected);
      });

      it("Formats expired items with formatKey correctly", () => {
        const pastUnixTime = moment().subtract(1, "day").unix();
        const item = { expires_in: pastUnixTime };

        const formatted = wrapper.vm.formatKey(item.expires_in);
        expect(formatted).toBe(true);
      });

      it("Formats items that will not expire with formatKey correctly", () => {
        const item = { expires_in: -1 };

        const formatted = wrapper.vm.formatKey(item.expires_in);
        expect(formatted).toBe(false);
      });
    });
  });
});
