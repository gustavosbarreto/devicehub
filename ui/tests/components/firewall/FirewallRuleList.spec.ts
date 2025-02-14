import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import FirewallRuleList from "@/components/firewall/FirewallRuleList.vue";
import { envVariables } from "@/envVariables";
import { router } from "@/router";
import { namespacesApi, rulesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type FirewallRuleListWrapper = VueWrapper<InstanceType<typeof FirewallRuleList>>;

describe("Firewall Rule List", () => {
  let wrapper: FirewallRuleListWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockFirewall: MockAdapter;

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
    name: "user",
    owner: "xxxxxxxx",
    tenant_id: "fake-tenant-data",
    members,
    max_devices: 3,
    devices_count: 3,
    devices: 2,
    created_at: "",
  };

  const authData = {
    status: "",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
    role: "owner",
  };

  const firewallRule = [
    {
      id: "1000",
      tenant_id: "00000000-0000-4000-0000-000000000000",
      priority: 1,
      action: "allow",
      active: true,
      source_ip: ".*",
      username: ".*",
      filter: {
        hostname: ".*",
      },
    },
    {
      id: "1001",
      tenant_id: "00000000-0000-4000-0000-000000000000",
      priority: 2,
      action: "allow",
      active: false,
      source_ip: ".*",
      username: ".*",
      filter: {
        hostname: ".*",
      },
    },
  ];

  const res = {
    data: firewallRule,
    headers: {
      "x-total-count": 2,
    },
  };

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockFirewall = new MockAdapter(rulesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockFirewall.onGet("http://localhost:3000/api/firewall/rules?page=1&per_page=10").reply(200, res);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("firewallRules/setFirewalls", res);

    wrapper = mount(FirewallRuleList, {
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

  it("Renders the Firewall List", async () => {
    expect(wrapper.find('[data-test="firewallRules-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-active"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-priority"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-action"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-source-ip"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-username"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-filter"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-actions"]').exists()).toBe(true);
  });
});
