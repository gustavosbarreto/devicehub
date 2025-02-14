import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { mfaApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import RecoveryHelper from "@/components/AuthMFA/RecoveryHelper.vue";

type RecoveryHelperWrapper = VueWrapper<InstanceType<typeof RecoveryHelper>>;

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

describe("Recovery Helper", () => {
  let wrapper: RecoveryHelperWrapper;
  const vuetify = createVuetify();

  let mock: MockAdapter;

  beforeEach(() => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    // Use fake timers and set a token in local storage
    vi.useFakeTimers();
    localStorage.setItem("token", "token");

    // Create a mock adapter for the mfaApi instance
    mock = new MockAdapter(mfaApi.getAxios());

    // Mount the MfaLogin component with necessary dependencies
    wrapper = mount(RecoveryHelper, {
      global: {
        plugins: [[store, key], vuetify, router],
      },
      attachTo: el,
    });
  });

  afterEach(() => {
    // Restore real timers and reset mocks after each test
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

  it("Renders the template with data", () => {
    const wrapper = new DOMWrapper(document.body);
    // Test if the component's template contains expected elements
    expect(wrapper.find('[data-test="card-dialog"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="card-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="alert"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="alert-second"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="checkbox"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="checkbox-recovery"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="disable-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("Disable MFA Authentication", async () => {
    // Mock the API response for MFA disable
    mock.onPut("http://localhost:3000/api/users/mfa/disable").reply(200);
    // Spy on Vuex store dispatch
    const mfaSpy = vi.spyOn(store, "dispatch");
    // Click the "Disable" button
    await wrapper.findComponent('[data-test="disable-btn"]').trigger("click");

    // Assert that the MFA disable action was dispatched
    expect(mfaSpy).toHaveBeenCalledWith("auth/disableMfa", { recovery_code: "" });
  });

  it("Disable MFA Authentication (fail)", async () => {
    // Mock the API response for MFA disable
    mock.onPut("http://localhost:3000/api/users/mfa/disable").reply(403);
    // Spy on Vuex store dispatch
    const mfaSpy = vi.spyOn(store, "dispatch");
    // Click the "Disable" button
    await wrapper.findComponent('[data-test="disable-btn"]').trigger("click");
    await flushPromises();
    // Assert that the MFA disable action was dispatched
    expect(mfaSpy).toHaveBeenCalledWith("snackbar/setSnackbarErrorDefault");
  });
});
