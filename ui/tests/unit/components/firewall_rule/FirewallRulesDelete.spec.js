import Vuex from 'vuex';
import { shallowMount, createLocalVue } from '@vue/test-utils';
import FirewallRulesDelete from '@/components/firewall_rules/FirewallRulesDelete';

describe('FirewallRulesDelete', () => {
  const localVue = createLocalVue();
  localVue.use(Vuex);

  let wrapper;
  const id = '5f1996c84d2190a22d5857bb';

  const store = new Vuex.Store({
    namespaced: true,
    state: {
    },
    getters: {
    },
    actions: {
      'firewallrules/remove': () => {
      },
    },
  });

  beforeEach(() => {
    wrapper = shallowMount(FirewallRulesDelete, {
      store,
      localVue,
      stubs: ['fragment'],
      props: { id },
    });
  });

  it('Is a Vue instance', () => {
    expect(wrapper).toBeTruthy();
  });
  it('Renders the component', () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
