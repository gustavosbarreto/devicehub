import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import SettingProfile from "@/components/Setting/SettingProfile.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type SettingProfileWrapper = VueWrapper<InstanceType<typeof SettingProfile>>;

describe("Settings Namespace", () => {
  let wrapper: SettingProfileWrapper;

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
    recovery_email: "test2@test.com",
    id: "xxxxxxxx",
    auth_methods: ["local", "saml"],
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  const session = true;

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
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("security/setSecurity", session);

    wrapper = mount(SettingProfile, {
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

  it("Renders all data-test elements", () => {
    const dataTests = [
      "account-profile-container",
      "account-profile-card",
      "profile-header",
      "user-icon",
      "profile-title",
      "profile-subtitle",
      "edit-profile-button",
      "profile-details-list",
      "name-field",
      "name-input",
      "username-field",
      "username-input",
      "email-field",
      "email-input",
      "recovery-email-field",
      "recovery-email-input",
      "mfa-card",
      "mfa-text",
      "switch-mfa",
      "delete-account",
      "delete-account-btn",
    ];

    dataTests.forEach((dataTest) => {
      const element = wrapper.find(`[data-test="${dataTest}"]`);
      expect(element.exists()).toBe(true);
    });
  });

  it("Successfully changes user data", async () => {
    mockUser.onPatch("http://localhost:3000/api/users").reply(200);

    await wrapper.findComponent('[data-test="edit-profile-button"]').trigger("click");
    await wrapper.findComponent('[data-test="name-input"]').setValue("test");
    await wrapper.findComponent('[data-test="username-input"]').setValue("test");
    await wrapper.findComponent('[data-test="email-input"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="recovery-email-input"]').setValue("test2@test.com");
    await flushPromises();

    const changeDataSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="save-changes-button"]').trigger("click");

    vi.runOnlyPendingTimers();

    await nextTick();
    await flushPromises();
    expect(changeDataSpy).toHaveBeenCalledWith("users/patchData", {
      email: "test@test.com",
      name: "test",
      recovery_email: "test2@test.com",
      username: "test",
    });
  });

  it("Fails changes user data", async () => {
    const changeUserData = {
      recovery_email: "test2@test.com",
      email: "test@test.com",
      name: "test",
      username: "test",
    };

    mockUser.onPatch("http://localhost:3000/api/users").reply(401);

    const changeDataSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="edit-profile-button"]').trigger("click");
    await wrapper.findComponent('[data-test="name-input"]').setValue("test");
    await wrapper.findComponent('[data-test="username-input"]').setValue("test");
    await wrapper.findComponent('[data-test="email-input"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="recovery-email-input"]').setValue("test2@test.com");
    await flushPromises();

    await wrapper.findComponent('[data-test="save-changes-button"]').trigger("click");

    vi.runOnlyPendingTimers();

    await nextTick();
    await flushPromises();
    expect(changeDataSpy).toHaveBeenCalledWith("users/patchData", changeUserData);
  });
});
