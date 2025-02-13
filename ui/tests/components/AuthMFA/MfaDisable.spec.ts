import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaDisable from "@/components/AuthMFA/MfaDisable.vue";
import { mfaApi, namespacesApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type MfaDisableWrapper = VueWrapper<InstanceType<typeof MfaDisable>>;

describe("MfaDisable", () => {
  const node = document.createElement("div");
  node.setAttribute("id", "app");
  document.body.appendChild(node);

  let wrapper: MfaDisableWrapper;

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
    wrapper = mount(MfaDisable, {
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

  it("Dialog opens", async () => {
    // Test if the dialog opens when the button is clicked
    wrapper.vm.showDialog = true;
    await flushPromises();
    // Check if the dialog element is not null
    expect(document.querySelector('[data-test="dialog"]')).not.toBeNull();
  });

  it("Renders the components", async () => {
    const dialog = new DOMWrapper(document.body);

    // Open the dialog
    wrapper.vm.showDialog = true;
    await flushPromises();

    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="sub-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="verification-code"]').exists()).toBe(true);
    expect(dialog.find('[data-test="verify-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="use-recovery-code-btn"]').exists()).toBe(true);
    wrapper.vm.el = 2;
    await flushPromises();
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="sub-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="recovery-code"]').exists()).toBe(true);
    expect(dialog.find('[data-test="recover-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="send-email-btn"]').exists()).toBe(true);
    wrapper.vm.el = 3;
    await flushPromises();
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="sub-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("Disable MFA Authentication using TOPT Code", async () => {
    const dialog = new DOMWrapper(document.body);
    // Test the scenario where MFA authentication is successfully disabled
    wrapper.vm.showDialog = true;
    await flushPromises();
    // Spy on Vuex store dispatch
    const mfaSpy = vi.spyOn(store, "dispatch");
    // Mock the API response for MFA disable
    mock.onPut("http://localhost:3000/api/user/mfa/disable").reply(200);
    await wrapper.findComponent('[data-test="verification-code"]').setValue("123456");
    // Click the "Disable" button
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");

    // Assert that the MFA disable action was dispatched
    expect(mfaSpy).toHaveBeenCalledWith("auth/disableMfa", { code: "123456" });
    expect(dialog.find('[data-test="alert-message"]').exists()).toBe(false);
  });

  it("Disable MFA Authentication using TOPT Code (Fail)", async () => {
    const dialog = new DOMWrapper(document.body);

    // Test the scenario where MFA authentication is successfully disabled
    wrapper.vm.showDialog = true;
    await flushPromises();
    // Spy on Vuex store dispatch
    const mfaSpy = vi.spyOn(store, "dispatch");
    // Mock the API response for MFA disable
    mock.onPut("http://localhost:3000/api/user/mfa/disable").reply(403);
    await wrapper.findComponent('[data-test="verification-code"]').setValue("123456");
    // Click the "Disable" button
    await wrapper.findComponent('[data-test="verify-btn"]').trigger("click");
    await flushPromises();
    // Assert that the MFA disable action was dispatched
    expect(mfaSpy).toHaveBeenCalledWith("auth/disableMfa", { code: "123456" });
    expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
  });

  it("Disable MFA Authentication using Recovery Code", async () => {
    // Test the scenario where MFA authentication is successfully disabled
    wrapper.vm.showDialog = true;
    wrapper.vm.el = 2;
    await flushPromises();
    // Spy on Vuex store dispatch
    const mfaSpy = vi.spyOn(store, "dispatch");
    // Mock the API response for MFA disable
    mock.onPut("http://localhost:3000/api/user/mfa/disable").reply(200);
    await wrapper.findComponent('[data-test="recovery-code"]').setValue("RMS32SAK521A");
    // Click the "Disable" button
    await wrapper.findComponent('[data-test="recover-btn"]').trigger("click");
    await flushPromises();

    // Assert that the MFA disable action was dispatched
    expect(mfaSpy).toHaveBeenCalledWith("auth/disableMfa", { recovery_code: "RMS32SAK521A" });
  });

  it("Disable MFA Authentication using Recovery Code (Fail)", async () => {
    const dialog = new DOMWrapper(document.body);

    // Test the scenario where MFA authentication is successfully disabled
    wrapper.vm.showDialog = true;
    wrapper.vm.el = 2;
    await flushPromises();
    // Spy on Vuex store dispatch
    const mfaSpy = vi.spyOn(store, "dispatch");
    // Mock the API response for MFA disable
    mock.onPut("http://localhost:3000/api/user/mfa/disable").reply(403);
    await wrapper.findComponent('[data-test="recovery-code"]').setValue("RMS32SAK521A");
    // Click the "Disable" button
    await wrapper.findComponent('[data-test="recover-btn"]').trigger("click");

    // Assert that the MFA disable action was dispatched
    expect(mfaSpy).toHaveBeenCalledWith("auth/disableMfa", { recovery_code: "RMS32SAK521A" });
    expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
  });

  it("Send the disable codes on the users mail", async () => {
    localStorage.setItem("email", "test@test.com");
    // Test the scenario where MFA authentication is successfully disabled
    wrapper.vm.showDialog = true;
    wrapper.vm.el = 2;
    await flushPromises();
    // Spy on Vuex store dispatch
    const mfaSpy = vi.spyOn(store, "dispatch");
    // Mock the API response for MFA disable
    mock.onPost("http://localhost:3000/api/user/mfa/reset").reply(200);
    // Click the "Disable" button
    await wrapper.findComponent('[data-test="send-email-btn"]').trigger("click");

    // Assert that the MFA disable action was dispatched
    expect(mfaSpy).toHaveBeenCalledWith("auth/reqResetMfa", "test@test.com");
  });
  it("Send the disable codes on the users mail", async () => {
    const dialog = new DOMWrapper(document.body);

    localStorage.setItem("email", "test@test.com");
    // Test the scenario where MFA authentication is successfully disabled
    wrapper.vm.showDialog = true;
    wrapper.vm.el = 2;
    await flushPromises();
    // Spy on Vuex store dispatch
    const mfaSpy = vi.spyOn(store, "dispatch");
    // Mock the API response for MFA disable
    mock.onPost("http://localhost:3000/api/user/mfa/reset").reply(403);
    // Click the "Disable" button
    await wrapper.findComponent('[data-test="send-email-btn"]').trigger("click");

    // Assert that the MFA disable action was dispatched
    expect(mfaSpy).toHaveBeenCalledWith("auth/reqResetMfa", "test@test.com");
    expect(dialog.find('[data-test="alert-message"]').exists()).toBe(true);
  });
});
