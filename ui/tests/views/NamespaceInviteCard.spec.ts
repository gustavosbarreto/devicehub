import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import NamespaceInviteCard from "@/views/NamespaceInviteCard.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type NamespaceInviteCardWrapper = VueWrapper<InstanceType<typeof NamespaceInviteCard>>;

let wrapper: NamespaceInviteCardWrapper;

const vuetify = createVuetify();

let mockNamespace: MockAdapter;

let mockUser: MockAdapter;

const members = [
  {
    id: "507f1f77bcf86cd799439011",
    username: "test",
    role: "owner",
  },
];

const namespaceData = {
  name: "test",
  owner: "test",
  tenant_id: "fake-tenant",
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

describe("Namespace Invite Dialog (Invalid User)", () => {
  beforeEach(async () => {
    // eslint-disable-next-line vue/max-len
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant");

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

    wrapper = mount(NamespaceInviteCard, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Displays appropriate error alert when user is not valid", async () => {
    localStorage.removeItem("id");

    await nextTick();

    expect(wrapper.vm.errorAlert).toBe("You aren't logged in the account meant for this invitation.");
    expect(wrapper.find('[data-test="decline-btn"]').exists()).toBe(true);
  });
});

describe("Namespace Invite Dialog", () => {
  beforeEach(async () => {
    // eslint-disable-next-line vue/max-len
    await router.push({ query: { "user-id": "507f1f77bcf86cd799439011" } });
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant");
    localStorage.setItem("id", "507f1f77bcf86cd799439011");

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

    wrapper = mount(NamespaceInviteCard, {
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

  it("Renders dialog elements with correct data-test attributes", async () => {
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="message"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="actions"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="decline-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="spacer"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="accept-btn"]').exists()).toBe(true);
  });

  it("Displays the correct title and message", () => {
    expect(wrapper.find('[data-test="title"]').text()).toBe(wrapper.vm.title);
    expect(wrapper.find('[data-test="message"]').text()).toBe(wrapper.vm.message);
  });

  it("Calls close method when decline button is clicked", async () => {
    const closeSpy = vi.spyOn(wrapper.vm, "close");
    await wrapper.findComponent('[data-test="decline-btn"]').trigger("click");

    await flushPromises();
    expect(closeSpy).toHaveBeenCalled();
  });

  it("Calls acceptInvite method when Accept Invitation button is clicked", async () => {
    // eslint-disable-next-line vue/max-len
    const acceptSpy = vi.spyOn(wrapper.vm, "acceptInvite");
    await flushPromises();
    await wrapper.findComponent('[data-test="accept-btn"]').trigger("click");

    await flushPromises();
    expect(acceptSpy).toHaveBeenCalled();
  });

  it("Handles error state correctly", async () => {
    // Simulate an error
    wrapper.vm.handleInviteError({ response: { status: 400 } });
    await nextTick();
    // eslint-disable-next-line vue/max-len
    expect(wrapper.find('[data-test="title"]').text()).toBe("Invite Accept Error");
    // eslint-disable-next-line vue/max-len
    expect(wrapper.find('[data-test="message"]').text()).toBe("An unexpected error occurred. Please try again later.");
    expect(wrapper.find('[data-test="accept-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="decline-btn"]').exists()).toBe(true);
  });
});
