import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { VLayout } from "vuetify/components";
import { namespacesApi, usersApi } from "@/api/http";
import AppBar from "@/components/AppBar/AppBar.vue";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";

const Component = {
  template: "<v-layout><AppBar /></v-layout>",
};

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

describe("AppBar Component", () => {
  let wrapper: VueWrapper<unknown>;
  const vuetify = createVuetify();
  envVariables.isCloud = true;

  beforeEach(async () => {
    window.matchMedia = vi.fn().mockImplementation((query) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));

    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    store.commit("auth/userInfo", { tenant: "fake-tenant-data" });

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    wrapper = mount(Component, {
      global: {
        plugins: [[store, key], vuetify, router],
        components: {
          "v-layout": VLayout,
          AppBar,
        },
      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    wrapper.unmount();
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

  it("Renders internal components", async () => {
    expect(wrapper.find('[data-test="app-bar"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="menu-toggle"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="breadcrumbs"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="support-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="user-menu-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="user-icon"]').exists()).toBe(true);
  });

  it("Opens the ShellHub help page when the support button is clicked", async () => {
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data/support").reply(200, { identifier: "fake-identifier" });

    const drawer = wrapper.findComponent(AppBar);

    const openShellhubHelpMock = vi.spyOn(drawer.vm, "openShellhubHelp");
    const supportBtn = wrapper.find('[data-test="support-btn"]');

    await supportBtn.trigger("click");
    expect(openShellhubHelpMock).toHaveBeenCalled();
  });

  it("Renders the logout btn", async () => {
    const drawer = wrapper.findComponent(AppBar);

    const userMenuBtn = drawer.find('[data-test="user-menu-btn"]');
    expect(userMenuBtn.exists()).toBe(true);

    await userMenuBtn.trigger("click");

    const logoutItem = drawer.findComponent('[data-test="Logout"]');
    expect(logoutItem.exists()).toBe(true);
  });

  it("Displays the correct breadcrumb titles", () => {
    const drawer = wrapper.findComponent(AppBar);

    const breadcrumbItems = drawer.findAll('[data-test="breadcrumbs"] v-breadcrumbs-item');

    const expectedBreadcrumbs = drawer.vm.breadcrumbItems;
    breadcrumbItems.forEach((item, index) => {
      expect(item.text()).toBe(expectedBreadcrumbs[index].title);
    });
  });

  it("Opens the default ShellHub support URL if identifier is not set", async () => {
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data/support").reply(200, { identifier: "" });

    const drawer = wrapper.findComponent(AppBar);

    const windowOpenMock = vi.spyOn(window, "open");

    await drawer.vm.openShellhubHelp();

    expect(windowOpenMock).toHaveBeenCalledWith(
      "https://github.com/shellhub-io/shellhub/issues/new/choose",
      "_blank",
    );
  });

  it("Uses Chatwoot if identifier is set", async () => {
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data/support").reply(200, { identifier: "fake-identifier" });

    const drawer = wrapper.findComponent(AppBar);
    const supportBtn = wrapper.find('[data-test="support-btn"]');

    vi.spyOn(drawer.vm, "identifier", "get").mockReturnValue("mocked_identifier");

    const windowOpenMock = vi.spyOn(window, "open").mockImplementation(() => null);
    const storeDispatchMock = vi.spyOn(store, "dispatch");

    await supportBtn.trigger("click");

    expect(windowOpenMock).not.toHaveBeenCalled();
    expect(storeDispatchMock).toHaveBeenCalledWith("support/get", "fake-tenant-data");
  });
});
