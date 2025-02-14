import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import Welcome from "@/components/Welcome/Welcome.vue";
import { envVariables } from "@/envVariables";
import { router } from "@/router";
import { namespacesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type WelcomeWrapper = VueWrapper<InstanceType<typeof Welcome>>;

describe("Welcome", () => {
  const node = document.createElement("div");
  node.setAttribute("id", "app");
  document.body.appendChild(node);

  let wrapper: WelcomeWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

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

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(Welcome, {
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
    wrapper.setProps({ show: true });
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.el = 1;
    await flushPromises();
    expect(dialog.find('[data-test="step-counter"]').exists()).toBe(true);
    expect(dialog.find('[data-test="welcome-first-screen"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="first-click-btn"]').exists()).toBe(true);
    wrapper.vm.el = 2;
    await flushPromises();
    expect(dialog.find('[data-test="welcome-second-screen"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close2-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="back-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="waiting-message"]').exists()).toBe(true);
    wrapper.vm.el = 3;
    await flushPromises();
    expect(dialog.find('[data-test="welcome-third-screen"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close3-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="back2-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="accept-btn"]').exists()).toBe(true);
    wrapper.vm.el = 4;
    await flushPromises();
    expect(dialog.find('[data-test="welcome-fourth-screen"]').exists()).toBe(true);
    expect(dialog.find('[data-test="finish-btn"]').exists()).toBe(true);
  });
  it("Renders the next btn when the user setups a device", async () => {
    wrapper.setProps({ show: true });
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.el = 2;
    wrapper.vm.enable = true;
    await flushPromises();
    expect(dialog.find('[data-test="next-btn"]').exists()).toBe(true);
  });
  it("Should go to the previous step when goToPreviousStep is called", async () => {
    wrapper.vm.el = 2;

    wrapper.vm.goToPreviousStep();
    await flushPromises();

    expect(wrapper.vm.el).toBe(1);
  });

  it("Should go to the next step when goToNextStep is called", async () => {
    wrapper.vm.el = 1;
    wrapper.vm.goToNextStep();

    await flushPromises();

    expect(wrapper.vm.el).toBe(2);
  });
});
