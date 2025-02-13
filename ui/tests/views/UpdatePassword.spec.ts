import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import UpdatePassword from "../../src/views/UpdatePassword.vue";
import { usersApi } from "@/api/http";
import { store, key } from "../../src/store";
import { router } from "../../src/router";
import { envVariables } from "../../src/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type UpdatePasswordWrapper = VueWrapper<InstanceType<typeof UpdatePassword>>;
const uid = "testID";

describe("Update Password", () => {
  let wrapper: UpdatePasswordWrapper;
  const vuetify = createVuetify();

  let mock: MockAdapter;

  beforeEach(async () => {
    vi.useFakeTimers();
    envVariables.isCloud = true;

    mock = new MockAdapter(usersApi.getAxios());
    await router.push(`/update-password?id=${uid}&token=testtoken`);

    wrapper = mount(UpdatePassword, {
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
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sub-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="password-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="password-confirm-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="update-password-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="isCloud-card"]').exists()).toBe(true);
  });

  it("Calls the Update Password Method when the button is clicked", async () => {
    const requestData = {
      password: "12345678",
      token: "testtoken",
      id: uid,
    };

    mock.onPost(`http://localhost:3000/api/user/${uid}/update_password`).reply(200);

    const updatePasswordSpy = vi.spyOn(store, "dispatch");
    const routerPushSpy = vi.spyOn(router, "push");

    await wrapper.findComponent('[data-test="password-text"]').setValue("12345678");
    await wrapper.findComponent('[data-test="password-confirm-text"]').setValue("12345678");

    await wrapper.findComponent('[data-test="update-password-btn"]').trigger("click");

    await flushPromises();

    expect(updatePasswordSpy).toHaveBeenCalledWith("users/updatePassword", requestData);
    expect(routerPushSpy).toHaveBeenCalledWith({ name: "Login" });
  });

  it("Error in updating password", async () => {
    mock.onPost(`http://localhost:3000/api/user/${uid}/update_password`).reply(400);

    const updatePasswordSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="password-text"]').setValue("12345");
    await wrapper.findComponent('[data-test="password-confirm-text"]').setValue("12345");

    await wrapper.findComponent('[data-test="update-password-btn"]').trigger("click");

    await flushPromises();

    expect(updatePasswordSpy).toHaveBeenCalledWith("snackbar/showSnackbarErrorAction", "updating account");
  });
});
