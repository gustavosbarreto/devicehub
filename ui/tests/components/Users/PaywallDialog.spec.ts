import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import axios from "axios";
import { store, key } from "@/store";
import PaywallDialog from "@/components/User/PaywallDialog.vue";
import { router } from "@/router";
import { namespacesApi, devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

const members = [
  {
    id: "xxxxxxxx",
    username: "test",
    role: "owner",
  },
];

const namespaceData = {
  name: "test",
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

const stats = {
  registered_devices: 3,
  online_devices: 1,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

const cards = [
  {
    title: "ShellHub Cloud",
    features: [
      "Protection Against DDoS Attacks",
      "Session record and playback",
      "Managing Firewall Rules",
      "Secure remote communication",
    ],
    button: {
      link: "https://www.shellhub.io/pricing",
      label: "Pricing",
    },
  },
  {
    title: "ShellHub Enterprise",
    features: [
      "Dedicated server for each customer",
      "Supports up to thousands of devices",
      "Reduced maintenance cost",
    ],
    button: {
      link: "https://www.shellhub.io/pricing",
      label: "Get a quote",
    },
  },
];

describe("PaywallDialog", async () => {
  let wrapper: VueWrapper<InstanceType<typeof PaywallDialog>>;

  const vuetify = createVuetify();

  let mock: MockAdapter;
  let mockNamespace: MockAdapter;
  let mockDevices: MockAdapter;

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);

    localStorage.setItem("tenant", "fake-tenant-data");

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());
    mock = new MockAdapter(axios);
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(PaywallDialog, {
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

  it("Renders the component table", async () => {
    wrapper.vm.dialog = true;
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="card-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="icon-crown"]').exists()).toBe(true);
    expect(dialog.find('[data-test="upgrade-heading"]').exists()).toBe(true);
    expect(dialog.find('[data-test="upgrade-description"]').exists()).toBe(true);

    // expect(dialog.find('[data-test="item-card-0"]').exists()).toBe(true);
    // expect(dialog.find('[data-test="item-card-1"]').exists()).toBe(true);

    // expect(dialog.find('[data-test="item-title-0"]').exists()).toBe(true);
    // expect(dialog.find('[data-test="item-title-1"]').exists()).toBe(true);

    // expect(dialog.find('[data-test="item-content-0"]').exists()).toBe(true);
    // expect(dialog.find('[data-test="item-content-1"]').exists()).toBe(true);

    // expect(dialog.find('[data-test="item-content-row-0-0"]').exists()).toBe(true);
    // expect(dialog.find('[data-test="item-content-row-0-1"]').exists()).toBe(true);
    // expect(dialog.find('[data-test="item-content-row-1-0"]').exists()).toBe(true);
    // expect(dialog.find('[data-test="item-content-row-1-1"]').exists()).toBe(true);

    // expect(dialog.find('[data-test="pricing-btn-0"]').exists()).toBe(true);
    // expect(dialog.find('[data-test="pricing-btn-1"]').exists()).toBe(true);

    // expect(dialog.find('[data-test="item-actions-0"]').exists()).toBe(true);
    // expect(dialog.find('[data-test="item-actions-1"]').exists()).toBe(true);

    expect(dialog.find('[data-test="card-actions"]').exists()).toBe(true);
    expect(dialog.find('[data-test="card-actions"]').exists()).toBe(true);

    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("Renders the component table with a successful request to get card infos", async () => {
    wrapper.vm.dialog = true;
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    mock.onGet("https://static.shellhub.io/premium-features.v1.json").reply(200, cards);

    store.commit("users/setPremiumContent", cards);

    await flushPromises();
    expect(dialog.find('[data-test="item-card-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-card-1"]').exists()).toBe(true);

    expect(dialog.find('[data-test="item-title-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-title-1"]').exists()).toBe(true);

    expect(dialog.find('[data-test="item-content-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-content-1"]').exists()).toBe(true);

    expect(dialog.find('[data-test="item-content-row-0-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-content-row-0-1"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-content-row-1-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-content-row-1-1"]').exists()).toBe(true);

    expect(dialog.find('[data-test="pricing-btn-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pricing-btn-1"]').exists()).toBe(true);

    expect(dialog.find('[data-test="item-actions-0"]').exists()).toBe(true);
    expect(dialog.find('[data-test="item-actions-1"]').exists()).toBe(true);
    expect(dialog.find('[data-test="no-link-available-btn"]').exists()).toBe(false);
  });

  it("Renders the component table with a successful request to get card infos", async () => {
    wrapper.vm.dialog = true;
    const dialog = new DOMWrapper(document.body);

    store.commit("users/setPremiumContent", []);

    await flushPromises();

    expect(dialog.find('[data-test="no-link-available-btn"]').exists()).toBe(true);
  });
});
