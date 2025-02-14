import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, afterEach, vi } from "vitest";
import { store, key } from "@/store";
import DeviceRename from "@/components/Devices/DeviceRename.vue";
import { envVariables } from "@/envVariables";
import { router } from "@/router";
import { namespacesApi, devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type DeviceRenameWrapper = VueWrapper<InstanceType<typeof DeviceRename>>;

describe("Device Rename", () => {
  let wrapper: DeviceRenameWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockDevice: MockAdapter;

  const device = {
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
    tenant_id: "00000000",
    last_seen: "2020-05-20T18:58:53.276Z",
    online: false,
    namespace: "user",
    status: "accepted",
  };

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
    max_devices: 3,
    devices_count: 3,
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

  beforeEach(async () => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevice = new MockAdapter(devicesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockDevice.onGet("http://localhost:3000/api/devices/a582b47a42d").reply(200, device);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("devices/setDevice", device);

    wrapper = mount(DeviceRename, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
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

  it("renders the component items", async () => {
    expect(wrapper.findComponent('[data-test="rename-icon"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="rename-title"]').exists()).toBe(true);

    wrapper.vm.showDialog = true;

    await flushPromises();
    expect(wrapper.findComponent('[data-test="deviceRename-card"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="text-title"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="rename-field"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="close-btn"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="rename-btn"]').exists()).toBe(true);
  });

  it("renames sucessfully a device", async () => {
    await wrapper.setProps({ uid: "a582b47a42d" });

    wrapper.vm.showDialog = true;

    await flushPromises();

    mockDevice.onPut("http://localhost:3000/api/devices/a582b47a42d").reply(200);

    const deviceSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="rename-field"]').setValue("renamed-device");
    await wrapper.findComponent('[data-test="rename-btn"]').trigger("click");

    await flushPromises();

    expect(deviceSpy).toHaveBeenCalledWith("devices/rename", {
      uid: "a582b47a42d",
      name: { name: "renamed-device" },
    });
  });

  it("renames sucessfully a device", async () => {
    await wrapper.setProps({ uid: "a582b47a42d" });

    wrapper.vm.showDialog = true;

    await flushPromises();

    mockDevice.onPut("http://localhost:3000/api/devices/a582b47a42d").reply(400);

    const deviceSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="rename-field"]').setValue("badly renamed device");
    await wrapper.findComponent('[data-test="rename-btn"]').trigger("click");

    await flushPromises();

    expect(deviceSpy).toHaveBeenCalledWith("devices/rename", {
      uid: "a582b47a42d",
      name: { name: "badly renamed device" },
    });
  });
});
