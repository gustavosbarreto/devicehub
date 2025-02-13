import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PrivateKeyDelete from "@/components/PrivateKeys/PrivateKeyDelete.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

type PrivateKeyDeleteWrapper = VueWrapper<InstanceType<typeof PrivateKeyDelete>>;

describe("Private Key Delete", () => {
  let wrapper: PrivateKeyDeleteWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

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
    settings: {
      session_record: true,
      connection_announcement: "",
    },
    max_devices: 3,
    devices_count: 3,
    created_at: "",
  };

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
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
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    wrapper = mount(PrivateKeyDelete, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("security/setSecurity", session);
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
    expect(wrapper.find('[data-test="privatekey-delete-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privatekey-delete-btn-title"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="privatekey-delete-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="privatekey-dialog-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="privatekey-dialog-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="privatekey-close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="privatekey-remove-btn"]').exists()).toBe(true);
  });

  it("Checks if the remove function updates the store on success", async () => {
    const storeSpy = vi.spyOn(store, "dispatch");
    await wrapper.setProps({ id: 1 });
    await wrapper.findComponent('[data-test="privatekey-delete-btn"]').trigger("click");
    await flushPromises();
    await wrapper.findComponent('[data-test="privatekey-remove-btn"]').trigger("click");
    expect(storeSpy).toHaveBeenCalledWith("privateKey/remove", 1);
  });
});
