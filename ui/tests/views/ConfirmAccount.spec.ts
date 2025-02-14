import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import ConfirmAccount from "../../src/views/ConfirmAccount.vue";
import { usersApi } from "@/api/http";
import { store, key } from "../../src/store";
import { router } from "../../src/router";
import { envVariables } from "../../src/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type ConfirmAccountWrapper = VueWrapper<InstanceType<typeof ConfirmAccount>>;
const username = "test";
describe("Login", () => {
  let wrapper: ConfirmAccountWrapper;
  const vuetify = createVuetify();

  let mock: MockAdapter;

  beforeEach(async () => {
    vi.useFakeTimers();
    envVariables.isCloud = true;
    await router.push(`/confirm-account?username=${username}`);

    // Create a mock adapter for the usersApi instance
    mock = new MockAdapter(usersApi.getAxios());

    wrapper = mount(ConfirmAccount, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    mock.reset();
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="subtitle"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="resendEmail-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="isCloud-card"]').exists()).toBe(true);
  });

  it("Resends an email to the user", async () => {
    const resendEmailSpy = vi.spyOn(store, "dispatch");

    mock.onPost("http://localhost:3000/api/user/resend_email").reply(200);
    await wrapper.findComponent('[data-test="resendEmail-btn"]').trigger("click");
    await flushPromises();

    expect(resendEmailSpy).toHaveBeenCalledWith("users/resendEmail", username);
  });

  it("Error case on resends an email to the user", async () => {
    const resendEmailSpy = vi.spyOn(store, "dispatch");

    mock.onPost("http://localhost:3000/api/user/resend_email").reply(400);
    await wrapper.findComponent('[data-test="resendEmail-btn"]').trigger("click");
    await flushPromises();

    expect(resendEmailSpy).toHaveBeenCalledWith("snackbar/showSnackbarErrorDefault", "resend email");
  });
});
