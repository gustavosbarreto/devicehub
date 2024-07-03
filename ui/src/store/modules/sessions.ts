import { Module } from "vuex";
import * as apiSession from "../api/sessions";
import { ISessions } from "@/interfaces/ISessions";
import { State } from "..";
import { ITerminalLog } from "@/interfaces/ITerminal";

export interface SessionsState {
  sessions: Array<ISessions>;
  session: Array<ITerminalLog>;
  numberSessions: number;
  page: number;
  perPage: number;
}

export const sessions: Module<SessionsState, State> = {
  namespaced: true,
  state: {
    sessions: [],
    session: {} as ITerminalLog[],
    numberSessions: 0,
    page: 1,
    perPage: 10,
  },

  getters: {
    list: (state) => state.sessions,
    get: (state) => state.session,
    getNumberSessions: (state) => state.numberSessions,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
  },

  mutations: {
    setSessions: (state, res) => {
      state.sessions = res.data;
      state.numberSessions = parseInt(res.headers["x-total-count"], 10);
    },

    setSession: (state, res) => {
      state.session = res.data;
    },

    updateSession: (state, res) => {
      const newSession = res.data;

      Object.keys(newSession).forEach((key) => {
        if (Object.prototype.hasOwnProperty.call(state.session, key)) {
          state.session[key] = newSession[key];
        } else {
          state.session = { ...state.session, [key]: newSession[key] };
        }
      });
    },

    setPagePerpage: (state, data) => {
      state.page = data.page;
      state.perPage = data.perPage;
    },

    resetPagePerpage: (state) => {
      state.page = 1;
      state.perPage = 10;
    },

    clearListSessions: (state) => {
      state.sessions = [];
      state.numberSessions = 0;
    },

    clearObjectSession: (state) => {
      state.session = [];
    },

    removeRecordedSession: (state) => {
      state.session = {
        ...state.session,
      };
    },
  },

  actions: {
    fetch: async (context, data) => {
      try {
        const res = await apiSession.fetchSessions(data.page, data.perPage);
        if (res.data.length) {
          context.commit("setPagePerpage", data);
          context.commit("setSessions", res);
          return res;
        }
        return false;
      } catch (error) {
        context.commit("clearListSessions");
        throw error;
      }
    },

    refresh: async (context) => {
      try {
        const res = await apiSession.fetchSessions(
          context.state.page,
          context.state.perPage,
        );
        context.commit("setSessions", res);
      } catch (error) {
        context.commit("clearListSessions");
        throw error;
      }
    },

    get: async (context, uid) => {
      try {
        const res = await apiSession.getSession(uid);
        context.commit("setSession", res);
      } catch (error) {
        context.commit("clearObjectSession");
        throw error;
      }
    },

    getLogSession: async (context, uid) => {
      try {
        const res = await apiSession.getLog(uid);
        context.commit("setSession", res);
      } catch (error) {
        context.commit("clearObjectSession");
        throw error;
      }
    },

    updateLogSession: async (context, uid) => {
      try {
        const res = await apiSession.getLog(uid);

        if (context.state.session.length === res.data.length) {
          throw new Error();
        }

        context.commit("setSession", res);
      } catch (error) {
        context.commit("clearObjectSession");
        throw error;
      }
    },

    resetPagePerpage: async (context) => {
      context.commit("resetPagePerpage");
    },

    close: async (context, session) => {
      try {
        await apiSession.closeSession(session);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    deleteSessionLogs: async (context, uid) => {
      try {
        await apiSession.deleteSessionLogs(uid);
        context.commit("removeRecordedSession");
      } catch (error) {
        console.error(error);
        throw error;
      }
    },
  },
};
