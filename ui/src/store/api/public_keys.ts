import http from "../helpers/http";
import { sshApi } from "../../api/http";
import { PublicKeyRequest, UpdatePublicKeyRequest } from "@/api/client/api";

export const postPublicKey = async (data : PublicKeyRequest) => sshApi.createPublicKey(data);

export const fetchPublicKeys = async (
  page : number,
  perPage: number,
  filter : string,
) => sshApi.getPublicKeys(filter, page, perPage);

export const getPublicKey = async (fingerprint : string) => http().get(`/sshkeys/public-keys/${fingerprint}`); // TODO

export const putPublicKey = async (data : { fingerprint: string } & UpdatePublicKeyRequest) => {
  sshApi.updatePublicKey(data.fingerprint, data);
};

export const removePublicKey = async (fingerprint : string) => sshApi.deletePublicKey(fingerprint);
