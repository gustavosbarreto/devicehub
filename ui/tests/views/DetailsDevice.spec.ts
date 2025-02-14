import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import { createRouter, createWebHistory } from "vue-router";
import DetailsDevice from "@/views/DetailsDevice.vue";
import { namespacesApi, usersApi, devicesApi } from "@/api/http";
import { store, key } from "@/store";
import { routes } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type DetailsDeviceWrapper = VueWrapper<InstanceType<typeof DetailsDevice>>;

describe("Details Device", () => {
  let wrapper: DetailsDeviceWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  let mockDevices: MockAdapter;

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
    id: "xxxxxxxx",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  const session = true;

  const res = {
    data: [namespaceData],
    headers: {
      "x-total-count": 1,
    },
  };

  const device = {
    uid: "123456",
    name: "00-00-00-00-00-01",
    identity: {
      mac: "00-00-00-00-00-01",
    },
    info: {
      id: "linux",
      pretty_name: "linux",
      version: "latest",
      arch: "amd64",
      platform: "docker",
    },
    public_key: "fake-public-key",
    tenant_id: "fake-tenant-data",
    last_seen: "",
    online: true,
    namespace: "dev",
    status: "accepted",
    status_updated_at: "",
    created_at: "",
    remote_addr: "192.168.0.10",
    position: {
      latitude: 0,
      longitude: 0,
    },
    tags: [],
    public_url: false,
    public_url_address: "",
    acceptable: false,
  };

  let router;

  beforeEach(async () => {
    router = createRouter({
      history: createWebHistory(),
      routes,
    });

    router.push("/devices/123456");

    await router.isReady();

    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");

    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());

    mockDevices.onGet("http://localhost:3000/api/devices/123456")
      .reply(200, device);
    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("namespaces/setNamespaces", res);
    store.commit("devices/setDevice", device);

    wrapper = mount(DetailsDevice, {
      global: {
        plugins: [[store, key], vuetify, [router], SnackbarPlugin],
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

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="deviceUid-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="deviceMac-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="devicePrettyName-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="deviceVersion-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="deviceTags-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="deviceConvertDate-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="terminalDialog-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="tunnel-list"]').exists()).toBe(false);
  });

  it("Renders the component when deviceIsEmpty is true", async () => {
    // Set device to empty object
    store.commit("devices/setDevice", {});
    await nextTick();
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component when device status is not accepted", async () => {
    // Set device status to 'pending'
    store.commit("devices/setDevice", { ...device, status: "pending" });
    await nextTick();
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component when device is offline", async () => {
    // Set device online status to false
    store.commit("devices/setDevice", { ...device, online: false });
    await nextTick();
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component when device has no tags", async () => {
    // Set device tags to empty array
    store.commit("devices/setDevice", { ...device, tags: [] });
    await nextTick();
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component when device has no last seen date", async () => {
    // Set device last_seen to empty string
    store.commit("devices/setDevice", { ...device, last_seen: "" });
    await nextTick();
    expect(wrapper.html()).toMatchSnapshot();
  });
});
