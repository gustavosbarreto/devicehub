import { Module } from "vuex";
import { State } from "./../index";
import * as apiUser from '../api/users';

export interface UsersState {
  statusUpdateAccountDialog: boolean;
  statusUpdateAccountDialogByDeviceAction: boolean;
}

export const users: Module<UsersState, State> = {
  namespaced: true,
  state: {
    statusUpdateAccountDialog: false,
    statusUpdateAccountDialogByDeviceAction: false,
  },

  getters: {
    statusUpdateAccountDialog: (state) => state.statusUpdateAccountDialog,
    statusUpdateAccountDialogByDeviceAction(state) {
      return state.statusUpdateAccountDialogByDeviceAction;
    },
  },

  mutations: {
    updateStatusUpdateAccountDialog(state, status) {
      state.statusUpdateAccountDialog = status;
    },

    updateStatusUpdateAccountDialogByDeviceAction(state, status) {
      state.statusUpdateAccountDialogByDeviceAction = status;
    },
  },

  actions: {
    async signUp(context, data) {
      await apiUser.signUp(data);
    },

    async patchData(context, data) {
      await apiUser.patchUserData(data);
    },

    async patchPassword(context, data) {
      await apiUser.patchUserPassword(data);
    },

    async resendEmail(context, username) {
      await apiUser.postResendEmail(username);
    },

    async recoverPassword(context, email) {
      await apiUser.postRecoverPassword(email);
    },

    async validationAccount(context, data) {
      await apiUser.postValidationAccount(data);
    },

    async updatePassword(context, data) {
      await apiUser.postUpdatePassword(data);
    },

    setStatusUpdateAccountDialog(context, status) {
      context.commit('updateStatusUpdateAccountDialog', status);
    },

    setStatusUpdateAccountDialogByDeviceAction(context, status) {
      context.commit('updateStatusUpdateAccountDialogByDeviceAction', status);
    },
  },
};
