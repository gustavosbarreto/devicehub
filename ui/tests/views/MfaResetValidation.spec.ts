import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import MfaResetValidation from "@/views/MfaResetValidation.vue";
import { mfaApi, namespacesApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type MfaResetValidationWrapper = VueWrapper<InstanceType<typeof MfaResetValidation>>;

describe("Validate Recovery Mail", () => {
  let wrapper: MfaResetValidationWrapper;

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

  beforeEach(async () => {
    await router.push("/reset-mfa?id=xxxxxx");

    vi.useFakeTimers();
    // Create a mock adapter for the mfaApi and namespacesApi instances
    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mock = new MockAdapter(mfaApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);

    // Commit auth and namespace data to the Vuex store
    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    // Mount the MfaDisable component with necessary dependencies
    wrapper = mount(MfaResetValidation, {
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
    await flushPromises();

    expect(wrapper.find('[data-test="verification-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="verification-subtitle"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="verification-success"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="verification-error"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="email-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="recovery-email-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="save-mail-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="back-to-login"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
  });

  it("Success on resetting the MFA", async () => {
    const storeSpy = vi.spyOn(store, "dispatch");
    mock.onPut("http://localhost:3000/api/user/mfa/reset/xxxxxx").reply(200);
    await wrapper.findComponent('[data-test="email-text"]').setValue("123");
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("1234");
    await wrapper.findComponent('[data-test="save-mail-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith("auth/resetMfa", { id: "xxxxxx", recovery_email_code: "1234", main_email_code: "123" });
  });

  it("Fails on resetting the MFA", async () => {
    const storeSpy = vi.spyOn(store, "dispatch");
    mock.onPut("http://localhost/api/user/mfa/reset/xxxxxx").reply(400);
    await wrapper.findComponent('[data-test="email-text"]').setValue("123");
    await wrapper.findComponent('[data-test="recovery-email-text"]').setValue("123");
    await wrapper.findComponent('[data-test="save-mail-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith("auth/resetMfa", { id: "xxxxxx", recovery_email_code: "123", main_email_code: "123" });
    expect(wrapper.find('[data-test="verification-error"]').exists()).toBe(true);
  });
});
