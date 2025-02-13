import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import UserDelete from "@/components/User/UserDelete.vue";
import { usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { INotificationsError } from "@/interfaces/INotifications";

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

type UserDeleteWrapper = VueWrapper<InstanceType<typeof UserDelete>>;

describe("User Delete", () => {
  let wrapper: UserDeleteWrapper;

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

    wrapper = mount(UserDelete, {
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
    expect(dialog.find('[data-test="user-delete-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="subtitle"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="delete-user-btn"]').exists()).toBe(true);
  });

  it("Successfully Delete User", async () => {
    mockUser.onDelete("http://localhost:3000/api/user").reply(200);

    const StoreSpy = vi.spyOn(store, "dispatch");

    wrapper.vm.show = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="delete-user-btn"]').trigger("click");

    await flushPromises();
    expect(StoreSpy).toHaveBeenCalledWith("auth/deleteUser");
  });

  it("Fails to add Delete User", async () => {
    mockUser.onDelete("http://localhost:3000/api/user").reply(400);

    const StoreSpy = vi.spyOn(store, "dispatch");

    wrapper.vm.show = true;
    await flushPromises();

    await wrapper.findComponent('[data-test="delete-user-btn"]').trigger("click");
    await flushPromises();
    expect(StoreSpy).toHaveBeenCalledWith(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.deleteAccount,
    );
  });
});
