import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import NamespaceEdit from "@/components/Namespace/NamespaceEdit.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { INotificationsError } from "@/interfaces/INotifications";

type NamespaceEditWrapper = VueWrapper<InstanceType<typeof NamespaceEdit>>;

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

describe("Namespace Edit", () => {
  let wrapper: NamespaceEditWrapper;

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

    wrapper = mount(NamespaceEdit, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
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
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.show = true;
    await flushPromises();
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="change-connection-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="connectionAnnouncement-text"]').exists()).toBe(true);
  });

  it("Successfully changes connection_announcement data", async () => {
    wrapper.vm.show = true;
    await flushPromises();
    const changeNamespaceData = {
      id: "fake-tenant-data",
      settings: {
        connection_announcement: "test",
      },
    };
    const response = {
      name: "test",
      owner: "test",
      tenant_id: "fake-tenant-data",
      members: [
        {
          id: "test",
          role: "owner",
        },
      ],
      settings: {
        session_record: true,
        connection_announcement: "test",
      },
      max_devices: 3,
      device_count: 3,
      created_at: "",
      billing: null,
    };
    mockNamespace.onPut("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, response);

    await wrapper.findComponent('[data-test="connectionAnnouncement-text"]').setValue("test");

    const changeDataSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="change-connection-btn"]').trigger("click");

    vi.runOnlyPendingTimers();

    await nextTick();
    await flushPromises();
    expect(changeDataSpy).toHaveBeenCalledWith("namespaces/put", changeNamespaceData);
  });

  it("Fails to change namespace data", async () => {
    wrapper.vm.show = true;
    await flushPromises();

    mockNamespace.onPut("http://localhost:3000/api/namespaces/fake-tenant-data").reply(403);

    await wrapper.findComponent('[data-test="connectionAnnouncement-text"]').setValue("test");

    const changeDataSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="change-connection-btn"]').trigger("click");

    vi.runOnlyPendingTimers();

    await nextTick();
    await flushPromises();
    expect(changeDataSpy).toHaveBeenCalledWith("snackbar/showSnackbarErrorAction", INotificationsError.namespaceEdit);
  });
});
