import { ratAPIBaseURL } from "./api";
import { Node } from "../types/node";
import axios from "axios";

export async function create(path: string, name: string) {
  let resp = await axios.post<Node>(`${ratAPIBaseURL()}/graph/node/${path}`, {
    name,
  });

  return resp.data;
}

export async function read(token: string, path: string) {
  let resp = await axios.get<Node>(`${ratAPIBaseURL()}/graph/node/${path}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return resp.data;
}

// delete is a reserved word in JavaScript, so we use remove instead.
export async function remove(path: string) {
  await axios.delete(`${ratAPIBaseURL()}/graph/node/${path}`);
}
