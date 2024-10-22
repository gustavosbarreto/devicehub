import { Module } from "vuex";
import { AxiosError } from "axios";
import * as apiAuth from "../api/auth";
import { IUserLogin } from "@/interfaces/IUserLogin";
import { State } from "..";

export interface AuthState {
  status: string;
  token: string;
  user: string;
  name: string;
  tenant: string;
  email: string;
  id: string;
  role: string;
  recoveryEmail: string;
  secret: string;
  linkMfa: string;
  mfa: boolean;
  recoveryCode: string;
  recoveryCodes: Array<number>;
  showRecoveryModal: boolean;
  loginTimeout: number;
  disableTimeout: number;
  mfaToken: string;
}

const localStorageSetItems = (data) => {
  localStorage.setItem("token", data.token || "");
  localStorage.setItem("user", data.user || "");
  localStorage.setItem("name", data.name || "");
  localStorage.setItem("tenant", data.tenant || "");
  localStorage.setItem("email", data.email || "");
  localStorage.setItem("id", data.id || "");
  localStorage.setItem("role", data.role || "");
  localStorage.setItem("mfa", data.mfa ? "true" : "false");
  localStorage.setItem("recovery_email", data.recovery_email || "");
};

const localStorageRemoveItems = () => {
  const keys = ["token", "user", "name", "tenant", "email", "id", "role", "mfa", "recovery_email"];
  keys.forEach((key) => localStorage.removeItem(key));
};

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
  },

  mutations: {
    authRequest(state) {
      state.status = "loading";
    },
    authSuccess(state, data) {
      Object.assign(state, {
        status: "success",
        token: data.token,
        user: data.user,
        name: data.name,
        tenant: data.tenant,
        email: data.email,
        id: data.id,
        role: data.role,
        mfa: data.mfa,
        recoveryEmail: data.recovery_email,
      });
    },
    authError(state) {
      state.status = "error";
    },
    logout(state) {
      Object.assign(state, {
        status: "",
        token: "",
        name: "",
        user: "",
        tenant: "",
        email: "",
        role: "",
        mfa: "",
      });
    },
    mfaEnabled(state, token) {
      state.mfa = true;
      localStorage.setItem("mfa", "true");
      state.mfaToken = token;
    },
    mfaDisable(state) {
      state.mfa = false;
      localStorage.setItem("mfa", "false");
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
      Object.assign(state, {
        linkMfa: data.link,
        secret: data.secret,
        recoveryCodes: data.codes,
        mfa: data.mfa,
        token: data.token,
        user: data.user,
        name: data.name,
        tenant: data.tenant,
        email: data.email,
        id: data.id,
        role: data.role,
        recoveryEmail: data.recovery_email,
      });
    },
    setToken(state, data) {
      state.token = data;
    },
    accountRecoveryHelper(state) {
      state.showRecoveryModal = !state.showRecoveryModal;
    },
    setLoginTimeout(state, timeout) {
      state.loginTimeout = timeout;
    },
    setDisableTimeout(state, timeout) {
      state.disableTimeout = timeout;
    },
    setRecoveryCode(state, code) {
      state.recoveryCode = code;
    },
  },

  actions: {
    async login({ commit }, user: IUserLogin) {
      commit("authRequest");
      try {
        const resp = await apiAuth.login(user);
        localStorageSetItems(resp.data);
        commit("authSuccess", resp.data);
      } catch (error: unknown) {
        const typedErr = error as AxiosError;
        if (typedErr.response?.headers["x-mfa-token"]) {
          localStorage.setItem("mfa", "true");
          commit("mfaEnabled", typedErr.response?.headers["x-mfa-token"]);
          return;
        }
        commit("setLoginTimeout", typedErr.response?.headers["x-account-lockout"]);
        commit("authError");
        throw error;
      }
    },

    async loginToken({ commit }, token) {
      commit("authRequest");
      localStorage.setItem("token", token);
      try {
        const resp = await apiAuth.info();
        localStorageSetItems(resp.data);
        commit("authSuccess", resp.data);
      } catch (error) {
        commit("authError");
      }
    },

    async disableMfa({ commit }, data) {
      await apiAuth.disableMfa(data);
      commit("mfaDisable");
    },

    async enableMfa({ commit }, data) {
      const resp = await apiAuth.enableMFA(data);
      if (resp.status === 200) {
        commit("mfaEnabled");
      }
    },

    async validateMfa({ commit }, data) {
      const resp = await apiAuth.validateMFA(data);
      if (resp.status === 200) {
        localStorageSetItems(resp.data);
        commit("authSuccess", resp.data);
      }
    },

    async generateMfa({ commit }) {
      const resp = await apiAuth.generateMfa();
      if (resp.status === 200) {
        commit("mfaGenerateInfo", resp.data);
      }
    },

    async getUserInfo({ commit }) {
      try {
        const resp = await apiAuth.info();
        if (resp.status === 200) {
          commit("userInfo", resp.data);
        }
      } catch (error) {
        commit("authError");
      }
    },

    async recoverLoginMfa({ commit }, data) {
      const resp = await apiAuth.validateRecoveryCodes(data);
      if (resp.status === 200) {
        localStorageSetItems(resp.data);
        commit("authSuccess", resp.data);
        commit("setToken", resp.data.token);
        commit("accountRecoveryHelper");
        commit("setDisableTimeout", resp.headers["x-expires-at"]);
      }
    },

    async reqResetMfa(_, data) {
      await apiAuth.reqResetMfa(data);
    },

    async resetMfa({ commit }, data) {
      const resp = await apiAuth.resetMfa(data);
      localStorageSetItems(resp.data);
      commit("authSuccess", resp.data);
    },

    logout({ commit }) {
      commit("logout");
      localStorageRemoveItems();
    },

    changeUserData({ commit }, data) {
      localStorage.setItem("name", data.name);
      localStorage.setItem("user", data.username);
      localStorage.setItem("email", data.email);
      localStorage.setItem("recovery_email", data.recoveryEmail);
      commit("changeData", data);
    },

    changeRecoveryEmail({ commit }, data) {
      localStorage.setItem("recovery_email", data);
      commit("changeRecoveryEmail", data);
    },

    setShowWelcomeScreen(_, tenantID: string) {
      const namespacesWelcome = JSON.parse(localStorage.getItem("namespacesWelcome") || "{}");
      namespacesWelcome[tenantID] = true;
      localStorage.setItem("namespacesWelcome", JSON.stringify(namespacesWelcome));
    },

    setDisableTokenTimeout({ commit }, timeout) {
      commit("setDisableTimeout", timeout);
    },
  },
};
