import Vuex from 'vuex';
import { mount, createLocalVue } from '@vue/test-utils';
import DeviceRejectedList from '@/components/device/DeviceRejectedList';
import Vuetify from 'vuetify';

describe('DeviceRejectedList', () => {
  const localVue = createLocalVue();
  const vuetify = new Vuetify();
  localVue.filter('moment', () => {});
  localVue.use(Vuex);

  let wrapper;
  let wrapper2;

  const numberDevices = 1;
  const devices = [
    {
      uid: '2378hj238',
      name: '37-23-hf-1c',
      identity: {
        mac: '00:00:00:00:00:00',
      },
      info: {
        id: 'linuxmint',
        pretty_name: 'Linux Mint 20.0',
        version: '',
      },
      public_key: '---pub_key---',
      tenant_id: '8490393000',
      last_seen: '2020-05-22T18:58:53.276Z',
      online: true,
      namespace: 'user',
      status: 'rejected',
    },
  ];
  const owner = true;

  const store = new Vuex.Store({
    namespaced: true,
    state: {
      devices,
      numberDevices,
      owner,
    },
    getters: {
      'devices/list': (state) => state.devices,
      'devices/getNumberDevices': (state) => state.numberDevices,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'modals/showAddDevice': () => {
      },
      'devices/fetch': () => {
      },
      'devices/rename': () => {
      },
      'devices/resetListDevices': () => {
      },
      'stats/get': () => {
      },
    },
  });

  const store2 = new Vuex.Store({
    namespaced: true,
    state: {
      devices,
      numberDevices,
      owner: false,
    },
    getters: {
      'devices/list': (state) => state.devices,
      'devices/getNumberDevices': (state) => state.numberDevices,
      'namespaces/owner': (state) => state.owner,
    },
    actions: {
      'modals/showAddDevice': () => {
      },
      'devices/fetch': () => {
      },
      'devices/rename': () => {
      },
      'devices/resetListDevices': () => {
      },
      'stats/get': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = mount(DeviceRejectedList, {
      store,
      localVue,
      stubs: ['fragment', 'router-link'],
      vuetify,
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it('Renders the template with data', () => {
    const dt = wrapper.find('[data-test="dataTable-field"]');
    const dataTableProps = dt.vm.$options.propsData;
    expect(dataTableProps.items).toHaveLength(numberDevices);
    expect(wrapper.find('[data-test="accept-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="remove-field"]').exists()).toBe(true);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.getListRejectedDevices).toEqual(devices);
    expect(wrapper.vm.getNumberRejectedDevices).toEqual(numberDevices);
  });
  it('Hides buttons for user not owner', () => {
    wrapper2 = mount(DeviceRejectedList, {
      store: store2,
      localVue,
      stubs: ['fragment', 'router-link'],
      vuetify,
    });
    expect(wrapper2.find('[data-test="accept-field"]').exists()).toBe(false);
    expect(wrapper2.find('[data-test="remove-field"]').exists()).toBe(false);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.getListRejectedDevices).toEqual(devices);
    expect(wrapper.vm.getNumberRejectedDevices).toEqual(numberDevices);
  });
});
