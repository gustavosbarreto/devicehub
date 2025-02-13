import { sessionsApi } from "../../api/http";
import { ISessions } from "@/interfaces/ISessions";

export const fetchSessions = async (page: number, perPage: number) => sessionsApi.getSessions(page, perPage);

export const getSession = async (uid: string) => sessionsApi.getSession(uid);

export const deleteSessionLogs = async (uid: string) => sessionsApi.clsoeSession(uid);

export const closeSession = async (
  session: ISessions,
) => sessionsApi.clsoeSession(session.uid, { device: session.device_uid });

export const getLog = async (uid: string) => sessionsApi.getSessionRecord(uid, 0);
