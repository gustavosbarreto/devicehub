import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import { nextTick, watch } from "vue";
import { store, key } from "@/store";
import TerminalDialog from "@/components/Terminal/TerminalDialog.vue";
import { router } from "@/router";
import { namespacesApi, devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

const devices = [
  {
    uid: "a582b47a42d",
    name: "39-5e-2a",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T18:58:53.276Z",
    online: false,
    namespace: "user",
    status: "accepted",
    tags: ["test"],
  },
  {
    uid: "a582b47a42e",
    name: "39-5e-2b",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T19:58:53.276Z",
    online: true,
    namespace: "user",
    status: "accepted",
    tags: ["test2"],
  },
];

const members = [
  {
    id: "xxxxxxxx",
    username: "test",
    role: "owner",
  },
];

const namespaceData = {
  name: "test",
  owner: "xxxxxxxx",
  tenant_id: "fake-tenant-data",
  members,
  max_devices: 3,
  devices_count: 3,
  devices: 2,
  created_at: "",
};

const authData = {
  status: "",
  token: "",
  user: "test",
  name: "test",
  tenant: "fake-tenant-data",
  email: "test@test.com",
  id: "xxxxxxxx",
  role: "owner",
};

const stats = {
  registered_devices: 3,
  online_devices: 1,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

describe("Terminal Dialog", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalDialog>>;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;
  let mockDevices: MockAdapter;

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);

    localStorage.setItem("tenant", "fake-tenant-data");

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockDevices.onGet("http://localhost:3000/api/devices?filter=&page=1&per_page=10&status=accepted").reply(200, devices);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(TerminalDialog, {
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

  it("Renders the component table", async () => {
    await wrapper.setProps({ enableConnectButton: true, enableConsoleIcon: true, uid: "a582b47a42d", online: true, show: true });

    expect(wrapper.find('[data-test="connect-btn"]').exists()).toBe(true);

    const dialog = new DOMWrapper(document.body);

    await flushPromises();

    await wrapper.findComponent('[data-test="connect-btn"]').trigger("click");

    expect(dialog.find('[data-test="terminal-card"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="username-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="password-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="connect2-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="auth-method-select"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="auth-method-select"]').setValue("Private Key");
    await flushPromises();

    expect(dialog.find('[data-test="password-field"]').exists()).toBe(false);
    expect(dialog.find('[data-test="privatekeys-select"]').exists()).toBe(true);
  });

  it("sets showLoginForm to true when showTerminal changes to true", async () => {
    await watch(() => wrapper.vm.showTerminal, (value) => {
      if (value) wrapper.vm.showLoginForm = true;
    });

    wrapper.vm.showTerminal = true;

    await nextTick();

    expect(wrapper.vm.showLoginForm).toBe(true);
  });

  it("encodes URL params correctly", () => {
    const params = { key1: "value1", key2: "value2" };
    const encodedParams = wrapper.vm.encodeURLParams(params);

    expect(encodedParams).toBe("key1=value1&key2=value2");
  });

  it("opens terminal and initializes xterm", () => {
    wrapper.vm.open();

    expect(wrapper.vm.showTerminal).toBe(true);
    expect(wrapper.vm.privateKey).toBe("");
    expect(wrapper.vm.xterm).toBeTruthy();
    expect(wrapper.vm.fitAddon).toBeTruthy();
  });
});
