import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import FirewallRuleAdd from "@/components/firewall/FirewallRuleAdd.vue";
import { envVariables } from "@/envVariables";
import { router } from "@/router";
import { namespacesApi, rulesApi, tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { INotificationsError } from "@/interfaces/INotifications";

type FirewallRuleAddWrapper = VueWrapper<InstanceType<typeof FirewallRuleAdd>>;

describe("Firewall Rule Add", () => {
  const node = document.createElement("div");
  node.setAttribute("id", "app");
  document.body.appendChild(node);

  let wrapper: FirewallRuleAddWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockFirewall: MockAdapter;

  let mockTags: MockAdapter;

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

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockFirewall = new MockAdapter(rulesApi.getAxios());
    mockTags = new MockAdapter(tagsApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockTags.onGet("http://localhost:3000/api/tags").reply(200);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(FirewallRuleAdd, {
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

  it("Renders the dialog open button and other key elements", async () => {
    const dialog = new DOMWrapper(document.body);

    expect(wrapper.find('[data-test="firewall-add-rule-btn"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="firewall-add-rule-btn"]').trigger("click");

    expect(dialog.find('[data-test="firewall-rule-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-status"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-priority"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-policy"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-source-ip"]').exists()).toBe(true);
    expect(dialog.find('[data-test="username-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="device-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-cancel"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-save-btn"]').exists()).toBe(true);
  });

  it("Conditional rendering components", async () => {
    const dialog = new DOMWrapper(document.body);

    wrapper.vm.choiceIP = "ipDetails";
    wrapper.vm.choiceUsername = "username";
    wrapper.vm.choiceFilter = "tags";

    await wrapper.findComponent('[data-test="firewall-add-rule-btn"]').trigger("click");

    expect(dialog.find('[data-test="firewall-rule-source-ip-details"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-username-restriction"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tags-selector"]').exists()).toBe(true);
  });

  it("Conditional rendering components (Hostname)", async () => {
    const dialog = new DOMWrapper(document.body);

    wrapper.vm.choiceFilter = "hostname";

    await wrapper.findComponent('[data-test="firewall-add-rule-btn"]').trigger("click");

    expect(dialog.find('[data-test="firewall-rule-hostname-restriction"]').exists()).toBe(true);
  });

  it("Successful on adding firewall rules", async () => {
    const storeSpy = vi.spyOn(store, "dispatch");

    mockFirewall.onPost("http://localhost:3000/api/firewall/rules").reply(200);

    await wrapper.findComponent('[data-test="firewall-add-rule-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="firewall-rule-save-btn"]').trigger("click");

    expect(storeSpy).toBeCalledWith("firewallRules/post", {
      policy: "allow",
      priority: 0,
      status: "active",
      source_ip: ".*",
      username: ".*",
      filter: {
        hostname: ".*",
      },
    });
  });

  it("Fails on adding firewall rules", async () => {
    const storeSpy = vi.spyOn(store, "dispatch");

    mockFirewall.onPost("http://localhost:3000/api/firewall/rules").reply(400);

    await wrapper.findComponent('[data-test="firewall-add-rule-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="firewall-rule-save-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toBeCalledWith(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.firewallRuleCreating,
    );
  });
});
