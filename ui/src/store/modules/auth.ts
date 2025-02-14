import { Module } from "vuex";
import { AxiosError } from "axios";
import { useChatWoot } from "@productdevbook/chatwoot/vue";
import * as apiAuth from "../api/auth";
import * as apiNamespace from "../api/namespaces";
import { IUserLogin } from "@/interfaces/IUserLogin";
import { State } from "..";
import {
  configuration as apiConfiguration,
  reloadConfiguration as reloadApiConfiguration,
} from "@/api/http";

const { reset, toggle } = useChatWoot();
export interface AuthState {
  status: string;
  token: string;
  user: string;
  name: string;
  tenant: string;
  email: string;
  id: string;
  role: string;
  auth_methods: Array<string>,
  recoveryEmail: string,
  secret: string;
  linkMfa: string;
  mfa: boolean;
  recoveryCode: string,
  recoveryCodes: Array<number>,
  showRecoveryModal: boolean,
  loginTimeout: number,
  disableTimeout: number,
  mfaToken: string,
}

export const auth: Module<AuthState, State> = {
  namespaced: true,
  state: {
    status: "",
    token: localStorage.getItem("token") || "",
    user: localStorage.getItem("user") || "",
    name: localStorage.getItem("name") || "",
    tenant: localStorage.getItem("tenant") || "",
    email: localStorage.getItem("email") || "",
    id: localStorage.getItem("id") || "",
    role: localStorage.getItem("role") || "",
    recoveryEmail: "",
    secret: "",
    linkMfa: "",
    mfa: false,
    auth_methods: ["local"],
    recoveryCode: "",
    recoveryCodes: [],
    showRecoveryModal: false,
    loginTimeout: 0,
    disableTimeout: 0,
    mfaToken: "",
  },

  getters: {
    isLoggedIn: (state) => !!state.token,
    authStatus: (state) => state.status,
    stateToken: (state) => state.token,
    currentUser: (state) => state.user,
    currentName: (state) => state.name,
    tenant: (state) => state.tenant,
    email: (state) => state.email,
    id: (state) => state.id,
    role: (state) => state.role,
    secret: (state) => state.secret,
    recoveryEmail: (state) => state.recoveryEmail,
    link_mfa: (state) => state.linkMfa,
    isMfa: (state) => state.mfa,
    mfaToken: (state) => state.mfaToken,
    stateRecoveryCode: (state) => state.recoveryCode,
    recoveryCodes: (state) => state.recoveryCodes,
    showRecoveryModal: (state) => state.showRecoveryModal,
    getLoginTimeout: (state) => state.loginTimeout,
    getDisableTokenTimeout: (state) => state.disableTimeout,
    showForceRecoveryMail: (state) => !state.recoveryEmail && state.mfa,
    getAuthMethods: (state) => state.auth_methods,
  },

  mutations: {
    authRequest(state) {
      state.status = "loading";
    },

    mfaEnabled(state, data) {
      state.mfa = true;
      localStorage.setItem("mfa", "true");
      state.mfaToken = data;
    },

    mfaDisable(state) {
      state.mfa = false;
      localStorage.setItem("mfa", "false");
    },

    mfaToken(state, data) {
      state.token = data;
    },

    authSuccess(state, data) {
      state.status = "success";
      state.token = data.token;
      state.user = data.user;
      state.name = data.name;
      state.tenant = data.tenant;
      state.email = data.email;
      state.id = data.id;
      state.role = data.role;
      state.mfa = data.mfa;
      state.recoveryEmail = data.recovery_email;
      state.auth_methods = data.auth_methods;
      localStorage.setItem("recovery_email", data.recovery_email);
    },

    authError(state) {
      state.status = "error";
    },

    logout(state) {
      state.status = "";
      state.token = "";
      state.name = "";
      state.user = "";
      state.tenant = "";
      state.email = "";
      state.role = "";
      state.mfa = false;
    },

    changeData(state, data) {
      state.name = data.name;
      state.user = data.username;
      state.email = data.email;
      state.recoveryEmail = data.recovery_email;
    },

    changeRecoveryEmail(state, data) {
      state.recoveryEmail = data;
    },

    mfaGenerateInfo(state, data) {
      state.linkMfa = data.link;
      state.secret = data.secret;
      state.recoveryCodes = data.recovery_codes;
    },

    userInfo(state, data) {
      state.linkMfa = data.link;
      state.secret = data.secret;
      state.recoveryCodes = data.codes;
      state.mfa = data.mfa;
      state.token = data.token;
      state.user = data.user;
      state.name = data.name;
      state.tenant = data.tenant;
      state.email = data.email;
      state.id = data.id;
      state.role = data.role;
      state.mfa = data.mfa;
      state.recoveryEmail = data.recovery_email;
      if (data.auth_methods !== null && data.auth_methods) {
        state.auth_methods = data.auth_methods;
      }
      localStorage.setItem("recovery_email", data.recovery_email);
    },

    accountRecoveryHelper(state) {
      state.showRecoveryModal = !state.showRecoveryModal;
    },

    setLoginTimeout: (state, data) => {
      state.loginTimeout = data;
    },

    setDisableTimeout: (state, data) => {
      state.disableTimeout = data;
    },

    setToken(state, data) {
      state.token = data;
    },

    setRecoveryCode: (state, data) => {
      state.recoveryCode = data;
    },
  },

  actions: {
    async login(context, user: IUserLogin) {
      context.commit("authRequest");
      try {
        const resp = await apiAuth.login(user);

        localStorage.setItem("token", resp.data.token || "");
        localStorage.setItem("user", resp.data.user || "");
        localStorage.setItem("name", resp.data.name || "");
        localStorage.setItem("tenant", resp.data.tenant || "");
        localStorage.setItem("email", resp.data.email || "");
        localStorage.setItem("id", resp.data.id || "");
        localStorage.setItem("namespacesWelcome", JSON.stringify({}));
        localStorage.setItem("role", resp.data.role || "");
        localStorage.setItem("mfa", "false");
        context.commit("authSuccess", resp.data);
      } catch (error: unknown) {
        const typedErr = error as AxiosError;
        if (typedErr.response?.headers["x-mfa-token"]) {
          localStorage.setItem("mfa", "true");
          context.commit("mfaEnabled", typedErr.response?.headers["x-mfa-token"]);
          return;
        }
        context.commit("setLoginTimeout", typedErr.response?.headers["x-account-lockout"]);
        context.commit("authError");
        throw error;
      }
    },

    async loginToken(context, token) {
      context.commit("authRequest");

      apiConfiguration.accessToken = token;
      reloadApiConfiguration();

      try {
        const resp = await apiAuth.info();

        localStorage.setItem("token", resp.data.token || "");
        localStorage.setItem("user", resp.data.user ?? "");
        localStorage.setItem("name", resp.data.name ?? "");
        localStorage.setItem("tenant", resp.data.tenant ?? "");
        localStorage.setItem("id", resp.data.id ?? "");
        localStorage.setItem("email", resp.data.email ?? "");
        localStorage.setItem("namespacesWelcome", JSON.stringify({}));
        localStorage.setItem("role", resp.data.role ?? "");
        context.commit("authSuccess", resp.data);
      } catch (error) {
        context.commit("authError");
      }
    },

    async disableMfa(context, data) {
      await apiAuth.disableMfa(data);
      context.commit("mfaDisable");
    },

    async enableMfa(context, data) {
      const resp = await apiAuth.enableMFA(data);

      if (resp.status === 200) {
        context.commit("mfaEnabled");
      }
    },

    async validateMfa(context, data) {
      const resp = await apiAuth.validateMFA(data);

      if (resp.status === 200) {
        localStorage.setItem("user", resp.data.user || "");
        localStorage.setItem("name", resp.data.name || "");
        localStorage.setItem("tenant", resp.data.tenant || "");
        localStorage.setItem("email", resp.data.email || "");
        localStorage.setItem("id", resp.data.id || "");
        localStorage.setItem("namespacesWelcome", JSON.stringify({}));
        localStorage.setItem("role", resp.data.role || "");
        localStorage.setItem("token", resp.data.token || "");
        localStorage.setItem("mfa", "true");
        context.commit("authSuccess", resp.data);
      }
    },

    enterInvitedNamespace: async (context, data) => {
      try {
        localStorage.removeItem("role");

        const res = await apiNamespace.tenantSwitch(data);
        if (res.status === 200) {
          localStorage.setItem("token", res.data.token || "");
          localStorage.setItem("tenant", data.tenant_id);
          localStorage.setItem("role", res.data.role || "");
        }
        context.commit("authSuccess", res.data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async generateMfa(context) {
      const resp = await apiAuth.generateMfa();
      if (resp.status === 200) {
        context.commit("mfaGenerateInfo", resp.data);
      }
    },

    async getUserInfo(context) {
      try {
        const resp = await apiAuth.info();
        if (resp.status === 200) {
          context.commit("userInfo", resp.data);
        }
      } catch (error) {
        context.commit("authError");
      }
    },

    async recoverLoginMfa(context, data) {
      const resp = await apiAuth.validateRecoveryCodes(data);
      if (resp.status === 200) {
        localStorage.setItem("user", resp.data.user || "");
        localStorage.setItem("name", resp.data.name || "");
        localStorage.setItem("tenant", resp.data.tenant || "");
        localStorage.setItem("email", resp.data.email || "");
        localStorage.setItem("id", resp.data.id || "");
        localStorage.setItem("namespacesWelcome", JSON.stringify({}));
        localStorage.setItem("role", resp.data.role || "");
        localStorage.setItem("token", resp.data.token || "");
        localStorage.setItem("mfa", "true");
        context.commit("authSuccess", resp.data);
        context.commit("mfaToken", resp.data.token);
        context.commit("accountRecoveryHelper");
        context.commit("setDisableTimeout", resp.headers["x-expires-at"]);
      }
    },

    async reqResetMfa(context, data) {
      await apiAuth.reqResetMfa(data);
    },

    async resetMfa(context, data) {
      const resp = await apiAuth.resetMfa(data);
      localStorage.setItem("token", resp.data.token || "");
      localStorage.setItem("user", resp.data.user || "");
      localStorage.setItem("name", resp.data.name || "");
      localStorage.setItem("tenant", resp.data.tenant || "");
      localStorage.setItem("email", resp.data.email || "");
      localStorage.setItem("id", resp.data.id || "");
      localStorage.setItem("namespacesWelcome", JSON.stringify({}));
      localStorage.setItem("role", resp.data.role || "");
      localStorage.setItem("mfa", "false");
      context.commit("authSuccess", resp.data);
    },

    async deleteUser(context) {
      await apiAuth.deleteUser();
      context.dispatch("logout");
    },

    logout(context) {
      context.commit("logout");
      localStorage.removeItem("token");
      localStorage.removeItem("user");
      localStorage.removeItem("tenant");
      localStorage.removeItem("namespacesWelcome");
      localStorage.removeItem("noNamespace");
      localStorage.removeItem("email");
      localStorage.removeItem("id");
      localStorage.removeItem("name");
      localStorage.removeItem("role");
      localStorage.removeItem("mfa");
      localStorage.removeItem("recovery_email");
      const chatIsCreated = context.rootGetters["support/getCreatedStatus"];
      if (chatIsCreated) {
        toggle("close");
        reset();
      }
    },

    changeUserData(context, data) {
      localStorage.setItem("name", data.name);
      localStorage.setItem("user", data.username);
      localStorage.setItem("email", data.email);
      localStorage.setItem("recovery_email", data.recoveryEmail);
      context.commit("changeData", data);
    },

    changeRecoveryEmail(context, data) {
      localStorage.setItem("recovery_email", data);
      context.commit("changeRecoveryEmail", data);
    },

    setShowWelcomeScreen(context, tenantID: string) {
      localStorage.setItem("namespacesWelcome", JSON.stringify(
        Object.assign(
          JSON.parse(localStorage.getItem("namespacesWelcome") || "") || {},
          { ...{ [tenantID]: true } },
        ),
      ));
    },
  },
};
