import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import FirewallRules from "@/views/FirewallRules.vue";
import { namespacesApi, usersApi, rulesApi } from "@/api/http";
import { store, key } from "@/store";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";

type FirewallRulesWrapper = VueWrapper<InstanceType<typeof FirewallRules>>;

describe("Firewall Rules", () => {
  let wrapper: FirewallRulesWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  let mockRules: MockAdapter;

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

  const firewallRule = {
    data: [{
      priority: 1,
      action: "allow",
      active: true,
      filter: {
        hostname: ".*",
      },
      source_ip: ".*",
      username: ".*",
    }],
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
    mockRules = new MockAdapter(rulesApi.getAxios());

    mockRules.onGet("http://localhost:3000/api/firewall/rules?page=1&per_page=10").reply(200, firewallRule.data);
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("namespaces/setNamespaces", res);
    store.commit("firewallRules/setFirewalls", firewallRule);
    store.commit("firewallRules/setFirewall", firewallRule.data[0]);

    wrapper = mount(FirewallRules, {
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
    expect(wrapper.find('[data-test="firewall-rules"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="help-icon"]').exists()).toBe(true);
  });

  it("Toggles showHelp when help icon is clicked", async () => {
    const helpIcon = wrapper.find('[data-test="help-icon"]');
    await helpIcon.trigger("click");
    expect(wrapper.vm.showHelp).toBe(true);

    await helpIcon.trigger("click");
    expect(wrapper.vm.showHelp).toBe(false);
  });
});
