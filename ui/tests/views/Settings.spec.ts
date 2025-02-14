import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import Settings from "@/views/Settings.vue";
import { store, key } from "@/store";
import { router } from "@/router";

type SettingsWrapper = VueWrapper<InstanceType<typeof Settings>>;

describe("Settings View", () => {
  let wrapper: SettingsWrapper;

  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(Settings, {
      global: {
        plugins: [[store, key], vuetify, router],
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
});
