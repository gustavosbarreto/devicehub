import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import ChangePassword from "@/components/User/ChangePassword.vue";
import { usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

type ChangePasswordWrapper = VueWrapper<InstanceType<typeof ChangePassword>>;

describe("Change Password", () => {
  let wrapper: ChangePasswordWrapper;

  const vuetify = createVuetify();

  let mockUser: MockAdapter;

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant",
    email: "test@test.com",
    id: "507f1f77bcf86cd799439011",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  const session = true;

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant");
    envVariables.isCloud = true;

    mockUser = new MockAdapter(usersApi.getAxios());

    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("security/setSecurity", session);

    wrapper = mount(ChangePassword, {
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

  it("Renders components", async () => {
    wrapper.vm.show = true;
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="password-change-card"]').exists()).toBe(true);
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="password-input"]').exists()).toBe(true);
    expect(dialog.find('[data-test="new-password-input"]').exists()).toBe(true);
    expect(dialog.find('[data-test="confirm-new-password-input"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="change-password-btn"]').exists()).toBe(true);
  });

  it("Successfully Change Password", async () => {
    mockUser.onPatch("http://localhost:3000/api/users").reply(200);

    const StoreSpy = vi.spyOn(store, "dispatch");

    wrapper.vm.show = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="password-input"]').setValue("xxxxxx");
    await wrapper.findComponent('[data-test="new-password-input"]').setValue("x1x2x3");
    await wrapper.findComponent('[data-test="confirm-new-password-input"]').setValue("x1x2x3");
    await wrapper.findComponent('[data-test="change-password-btn"]').trigger("click");

    await flushPromises();
    expect(StoreSpy).toHaveBeenCalledWith("users/patchPassword", {
      name: "test",
      username: undefined,
      email: "test@test.com",
      recovery_email: undefined,
      currentPassword: "xxxxxx",
      newPassword: "x1x2x3",
    });
  });

  it("Fails to Change Password", async () => {
    mockUser.onPatch("http://localhost:3000/api/users").reply(403);

    const StoreSpy = vi.spyOn(store, "dispatch");

    wrapper.vm.show = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="password-input"]').setValue("xxxxxx");
    await wrapper.findComponent('[data-test="new-password-input"]').setValue("x1x2x3");
    await wrapper.findComponent('[data-test="confirm-new-password-input"]').setValue("x1x2x3");

    await wrapper.findComponent('[data-test="change-password-btn"]').trigger("click");
    await flushPromises();

    expect(StoreSpy).toHaveBeenCalledWith("users/patchPassword", {
      name: "test",
      username: undefined,
      email: "test@test.com",
      recovery_email: undefined,
      currentPassword: "xxxxxx",
      newPassword: "x1x2x3",
    });

    expect(StoreSpy).toHaveBeenCalledWith("snackbar/showSnackbarErrorDefault");
  });
});
