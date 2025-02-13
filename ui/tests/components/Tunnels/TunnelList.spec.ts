import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import { createRouter, createWebHistory } from "vue-router";
import { nextTick } from "vue";
import { store, key } from "@/store";
import TunnelList from "@/components/Tunnels/TunnelList.vue";
import { routes } from "@/router";
import { namespacesApi, devicesApi, tunnelApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type TunnelListWrapper = VueWrapper<InstanceType<typeof TunnelList>>;

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

const tunnelResponse = [{
  address: "9a8df9321368d567cfac8679cec7848c",
  namespace: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
  device: "13b0c8ea878e61ff849db69461795006a9594c8f6a6390ce0000100b0c9d7d0a",
  host: "127.0.0.1",
  port: 8080,
}];

describe("Tunnel List", () => {
  let wrapper: TunnelListWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;
  let mockDevices: MockAdapter;
  let mockTunnels: MockAdapter;

  let router;

  beforeEach(async () => {
    router = createRouter({
      history: createWebHistory(),
      routes,
    });

    router.push("/devices/fake-uid");

    await router.isReady();

    localStorage.setItem("tenant", "fake-tenant-data");

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());
    mockTunnels = new MockAdapter(tunnelApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockDevices.onGet("http://localhost:3000/api/devices?filter=&page=1&per_page=10&status=accepted").reply(200, devices);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);
    mockTunnels.onGet("http://localhost:3000/api/devices/fake-uid/tunnels").reply(200, tunnelResponse);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("devices/setDeviceChooserStatus", true);

    wrapper = mount(TunnelList, {
      global: {
        plugins: [[store, key], vuetify, [router], SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the table", () => {
    expect(wrapper.find('[data-test="device-tunnels-table"]').exists()).toBe(true);
  });

  it("Renders table headers", () => {
    const headers = wrapper.findAll('[data-test^="device-tunnels-header-"]');
    expect(headers.length).toBe(5);
    expect(headers[0].text()).toBe("Address");
    expect(headers[1].text()).toBe("Host");
    expect(headers[2].text()).toBe("Port");
    expect(headers[3].text()).toBe("Expiration Date");
    expect(headers[4].text()).toBe("Actions");
  });

  it("Renders table rows", async () => {
    const rows = wrapper.findAll('[data-test^="device-tunnel-row-"]');
    rows.forEach((row, index) => {
      expect(row.find("[data-test=\"device-tunnel-url\"]").exists()).toBe(true);
      expect(row.find("[data-test=\"device-tunnel-host\"]").text()).toBe(tunnelResponse[index].host);
      expect(row.find("[data-test=\"device-tunnel-port\"]").text()).toBe(`${tunnelResponse[index].port}`);
      expect(row.find("[data-test^=\"device-tunnel-delete-\"]").exists()).toBe(true);
    });
  });

  it("Renders empty state if no tunnels", async () => {
    await mockTunnels.onGet("http://localhost:3000/api/devices/fake-uid/tunnels").reply(200, []);
    await wrapper.vm.getTunnels();
    await nextTick();

    expect(wrapper.find('[data-test="device-tunnels-empty"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-tunnels-empty"]').text()).toContain("No data available");
  });
});
