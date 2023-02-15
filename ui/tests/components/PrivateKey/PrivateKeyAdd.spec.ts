import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import PrivateKeyAdd from "../../../src/components/PrivateKeys/PrivateKeyAdd.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const privateKey = {
  name: "",
  data: "",
};

const publicKey = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCxXq0NZ
tbRBJlyyW5LOTMuqsZj3pL+Q5UCSQxnEjMpnz6yL6ALTS+fpVLzHIQwfZ3p5kMGk
vAwXOwLuvkFpvQvbGsj7/kBov6zDeL7exdzPVvhVclsIU//aTm2ryT1898RFgEOm
2YDSsNteG4dYBe9SbNJIbezAg7lCAdKxsbZD05phX8NewGOcFolPk8kSuYqJ6lWB
/WWncLT8eXgP8Ew95rwug9Am3ApijuoD1j1RIb1LirF9xkNNg13DA9QYEFOO16XV
EIxIS1frW7Krh+3LP2W6Q5ISFRzGF7hxlWs9RRzB/SG2WxrOpeQAoDOLrt/fu3g7
sVL9pA32YbLgyAT`;

const privateKeyRSA = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAsV6tDWbW0QSZcsluSzkzLqrGY96S/kOVAkkMZxIzKZ8+si+g
C00vn6VS8xyEMH2d6eZDBpLwMFzsC7r5Bab0L2xrI+/5AaL+sw3i+3sXcz1b4VXJ
bCFP/2k5tq8k9fPfERYBDptmA0rDbXhuHWAXvUmzSSG3swIO5QgHSsbG2Q9OaYV/
DXsBjnBaJT5PJErmKiepVgf1lp3C0/Hl4D/BMPea8LoPQJtwKYo7qA9Y9USG9S4q
xfcZDTYNdwwPUGBBTjtel1RCMSEtX61uyq4ftyz9lukOSEhUcxhe4cZVrPUUcwf0
htlsazqXkAKAzi67f37t4O7FS/aQN9mGy4MgEwIDAQABAoIBACm+XnwI+AW5T2P0
hECv9ZvGFWrrtyygOzGOf5zCa8gf8mF9U+1U/SBViHAvBe1jowapapzheCXpuKQK
HRF3eYCvw4dxsujvs4Hwgrss/RfiGP2vcvg/3tP9r9eO4OQBwT4EL7uAV0HvFo9t
CH1hYDTsY4WSqek3UsoUWaL/pUzwKMijUgh2Dzj5o9AlNGWANu6txI1mIgHmwUvj
2kV7E4R1mGynSprdsW68V36viB/V9d82XGxd3tYhKojiS1Dir68mR2U8ld3728Pd
xU7o9x6NcWOtpTY1nS9MpufaYUTlp/chOXSd2RIY6JmtgbJcVTdE4rasfIAEnlAZ
XALqKAECgYEA4kl6ZfcwKtxebVyczMCk6QWOJtsJ6CT17w2oehQGSuSLXeidjkFe
bm04hUcN4Rm5iipUwDlA6JT8QoUgSG7Mjf8aDLv68FjXHxHjVvQaj0pg2I+1qADZ
bN6m5xaazqAShF5MN4zQQTnNHTp6AIXOSQhIpqKS/Bjf3FYw48pxCyMCgYEAyKjf
GnwiFJZN/q3s2mCmlEPblJ5mbXGCmIK/wjcoDST3+YrFi5VoWsHu0hRoZHtxIiaH
sjSj8f8hWaZJ+yTL/V6zAO93JMovmoYyClmGt5pl56pFT2B7VGDC5FU9bylzWF3g
HDdCTXOE72c5cOHnOddxVSBdD6GLC7Qe4CUVnlECgYBYNmSskywHyVhWMaA+gWrI
HA5KP2EhSidFRYHD9UJut6FMvn2NExaI3bMG4agbdDfMEKxxMuCGym18UQFAu1Cq
miPBixZL05Yo2oRRRV+FNG2EfqFGGO6pbjKKK1m16tjNGSWFEjOs+adoGX+t7Ht6
JOyNaRr7g4bhEgiFBEoFGQKBgF3XI+dl8CZCmJ0nR6JlGuIxzen2Hh7Gu/WJCBbS
5qcnB9UrAfGiYNg44/BZXOzJEgKPlFxR4+4Ti8w6SVTrQ37tn7crRkPtTk/svFA8
yBTrXwb1iU5y55pxWhOgjYeEEg5ccKehbB9+i8fONX3GF/Xj/Ht8FClwOe+yP9JB
ZZfRAoGBAJb08mFdb0Csbp+ed3LFznWINpXf2vlRKqIf+w8VOsEItbiB0r08AVdA
Tik8VkRWm9ZHnMeMRRg2sEsI8gfaEXwSfLfMi10fn9YuWC2GSt5z+lA52H/S1zU2
sGHPNn1H/cu7eM+nr9NxzJIT2CzKMHt5w4epp/UgkYFri4n2wDNS
-----END RSA PRIVATE KEY-----`;

const tests = [
  {
    description: "Button create private Key",
    props: {
      size: "default",
    },
    data: {
      privateKey,
      supportedKeys:
        "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
    },
    template: {
      "createKey-btn": true,
      "privateKeyFormDialog-card": false,
    },
    templateText: {
      "createKey-btn": "Add Private Key",
    },
  },
  {
    description: "Dialog opened",
    props: {
      size: "default",
    },
    data: {
      privateKey,
      supportedKeys:
        "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
    },
    template: {
      "createKey-btn": true,
      "privateKeyFormDialog-card": true,
      "text-title": true,
      "name-field": true,
      "data-field": true,
      "cancel-btn": true,
      "create-btn": true,
    },
    templateText: {
      "createKey-btn": "Add Private Key",
      "text-title": "New Private Key",
      "name-field": "",
      "data-field": "",
      "cancel-btn": "Cancel",
      "create-btn": "Create",
    },
  },
];

const store = createStore({
  state: {},
  getters: {
    "auth/role": () => "admin",
  },
  actions: {
    "privateKey/set": vi.fn(),
    "snackbar/showSnackbarSuccessNotRequest": vi.fn(),
    "snackbar/showSnackbarErrorNotRequest": vi.fn(),
  },
});

describe("PrivateKeyFormDialogAdd", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(PrivateKeyAdd, {
          global: {
            plugins: [[store, key], routes, vuetify],
          },
          props: {
            size: test.props.size,
          },
          shallow: true,
        });
      });

      ///////
      // Component Rendering
      //////

      it("Is a Vue instance", () => {
        expect(wrapper).toBeTruthy();
      });
      it("Renders the component", () => {
        expect(wrapper.html()).toMatchSnapshot();
      });

      ///////
      // Data checking
      //////

      it('Compare data with default value', () => {
        expect(wrapper.vm.name).toBe(test.data.privateKey.name);
        expect(wrapper.vm.privateKeyData).toBe(test.data.privateKey.data);
        expect(wrapper.vm.supportedKeys).toBe(test.data.supportedKeys);
      });

      it("Compare data with props", () => {
        expect(wrapper.vm.size).toBe(test.props.size);
      });

      //////
      // HTML validation
      //////

      // TODO

    });
  }); 
});
