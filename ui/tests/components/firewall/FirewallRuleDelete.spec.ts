import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import FirewallRuleDelete from "@/components/firewall/FirewallRuleDelete.vue";
import { envVariables } from "@/envVariables";
import { router } from "@/router";
import { namespacesApi, rulesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { INotificationsError } from "@/interfaces/INotifications";

type FirewallRuleDeleteWrapper = VueWrapper<InstanceType<typeof FirewallRuleDelete>>;

describe("Firewall Rule Delete", () => {
  const node = document.createElement("div");
  node.setAttribute("id", "app");
  document.body.appendChild(node);

  let wrapper: FirewallRuleDeleteWrapper;

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

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockFirewall = new MockAdapter(rulesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(FirewallRuleDelete, {
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

    expect(wrapper.find('[data-test="remove-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="remove-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-delete-dialog-btn"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="firewall-delete-dialog-btn"]').trigger("click");

    expect(dialog.find('[data-test="firewallRuleDelete-card"]').exists()).toBe(true);
    expect(dialog.find('[data-test="text-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="text-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="remove-btn"]').exists()).toBe(true);
  });

  it("Successful on removing firewall rules", async () => {
    await wrapper.setProps({ id: "1000" });

    const storeSpy = vi.spyOn(store, "dispatch");

    mockFirewall.onDelete("http://localhost:3000/api/firewall/rules/1000").reply(200);

    await wrapper.findComponent('[data-test="firewall-delete-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");

    expect(storeSpy).toBeCalledWith("firewallRules/remove", "1000");
  });

  it("Fails on removing firewall rules", async () => {
    await wrapper.setProps({ id: "1000" });

    const storeSpy = vi.spyOn(store, "dispatch");

    mockFirewall.onDelete("http://localhost:3000/api/firewall/rules/1000").reply(403);

    await wrapper.findComponent('[data-test="firewall-delete-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.firewallRuleDeleting,
    );
  });
});
