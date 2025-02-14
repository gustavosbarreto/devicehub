import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PrivateKeyEdit from "@/components/PrivateKeys/PrivateKeyEdit.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { INotificationsError, INotificationsSuccess } from "@/interfaces/INotifications";

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

type PrivateKeyEditWrapper = VueWrapper<InstanceType<typeof PrivateKeyEdit>>;

describe("Private Key Edit", () => {
  let wrapper: PrivateKeyEditWrapper;

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

  const mockObject = {
    name: "test-name",
    data: "test-data",
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

    wrapper = mount(PrivateKeyEdit, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      props: {
        keyObject: mockObject,
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
    expect(wrapper.find('[data-test="privatekey-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privatekey-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privatekey-edit-btn"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="privatekey-edit-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="name-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="private-key-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pk-edit-cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="pk-edit-save-btn"]').exists()).toBe(true);
  });

  it("Checks if the private key data is valid", async () => {
    await wrapper.vm.setPrivateKey();
    const privateKeyData = wrapper.vm.keyLocal;
    expect(privateKeyData).toBeDefined();
    expect(wrapper.vm.isValid).toBe(true);
  });

  it("Checks if the name field is valid", async () => {
    await flushPromises();
    const nameField = wrapper.vm.name;
    expect(nameField).toBeDefined();
  });

  it("Checks if the update function emits an update event", async () => {
    await wrapper.vm.update();
    expect(wrapper.emitted().update).toBeTruthy();
  });

  it("Checks if the edit function updates the store on success", async () => {
    const storeSpy = vi.spyOn(store, "dispatch");
    await wrapper.vm.setPrivateKey();
    const keySend = { name: wrapper.vm.name, data: wrapper.vm.keyLocal };
    await wrapper.vm.edit();
    expect(storeSpy).toHaveBeenCalledWith("privateKey/edit", keySend);
    expect(storeSpy).toHaveBeenCalledWith("snackbar/showSnackbarSuccessAction", INotificationsSuccess.privateKeyEditing);
  });

  it("Checks if the edit function handles error on failure", async () => {
    const storeSpy = vi.spyOn(store, "dispatch");
    await wrapper.vm.setPrivateKey();
    await wrapper.vm.edit();
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith("snackbar/showSnackbarErrorAction", INotificationsError.privateKeyEditing);
  });
});
