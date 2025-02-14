import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaForceRecoveryMail from "@/components/AuthMFA/MfaForceRecoveryMail.vue";
import { mfaApi, namespacesApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type MfaForceRecoveryMailWrapper = VueWrapper<InstanceType<typeof MfaForceRecoveryMail>>;

describe("Force Adding a Recovery Mail", () => {
  const node = document.createElement("div");
  node.setAttribute("id", "app");
  document.body.appendChild(node);

  let wrapper: MfaForceRecoveryMailWrapper;

  const vuetify = createVuetify();

  let mock: MockAdapter;

  let mockNamespace: MockAdapter;

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
    max_devices: 3,
    devices_count: 3,
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
    recovery_email: "recover@mail.com",
    role: "owner",
    mfa: true,
  };

  beforeEach(() => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    // Create a mock adapter for the mfaApi and namespacesApi instances
    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mock = new MockAdapter(mfaApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);

    // Commit auth and namespace data to the Vuex store
    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    // Mount the MfaDisable component with necessary dependencies
    wrapper = mount(MfaForceRecoveryMail, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    mock.reset();
  });

  it("Is a Vue instance", () => {
    // Test if the wrapper represents a Vue instance
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    // Test if the component renders as expected
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Data is defined", () => {
    // Test if component data is defined
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Renders the components", async () => {
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.dialog = true;
    await flushPromises();

    expect(dialog.find('[data-test="card-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="recovery-email-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="save-btn"]').exists()).toBe(true);
  });

  it("Adds a recovery mail", async () => {
    wrapper.vm.dialog = true;
    await flushPromises();
    const storeSpy = vi.spyOn(store, "dispatch");
    mock.onPut("http://localhost:3000/api/users/xxxxxxxx/data").reply(200);
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("test2@test.com");
    await wrapper.findComponent('[data-test="save-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith("users/patchData", { id: "xxxxxxxx", recovery_email: "test2@test.com" });
    expect(wrapper.vm.recoveryEmailError).toBe(undefined);
  });

  it("Adds a recovery mail (Fail)", async () => {
    wrapper.vm.dialog = true;
    await flushPromises();
    const storeSpy = vi.spyOn(store, "dispatch");
    mock.onPut("http://localhost:3000/api/users/xxxxxxxx/data").reply(400);
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("test");
    await wrapper.findComponent('[data-test="save-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith("users/patchData", { id: "xxxxxxxx", recovery_email: "test" });
    expect(wrapper.vm.recoveryEmailError).toBe("Please enter a valid email address");
  });

  it("Adds a recovery mail (Fail, Same Email)", async () => {
    wrapper.vm.dialog = true;
    await flushPromises();
    const storeSpy = vi.spyOn(store, "dispatch");
    mock.onPut("http://localhost:3000/api/users/xxxxxxxx/data").reply(400);
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="save-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith("users/patchData", { id: "xxxxxxxx", recovery_email: "test@test.com" });
    expect(wrapper.vm.recoveryEmailError).toBe("Recovery email must not be the same as your current email");
  });
});
