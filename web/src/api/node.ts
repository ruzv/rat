import { ratAPIBaseURL } from "./api";
import { Node } from "../types/node";
import axios from "axios";

export async function create(path: string, name: string) {
  const url = `${ratAPIBaseURL()}/graph/node/${path}`;

  let resp = await axios.post<Node>(url, { name });

  return resp.data;
}

export async function read(path: string) {
  let resp = await axios.get<Node>(`${ratAPIBaseURL()}/graph/node/${path}`);

  return resp.data;
}

// delete is a reserved word in JavaScript, so we use remove instead.
export async function remove(path: string) {
  await axios.delete(`${ratAPIBaseURL()}/graph/node/${path}`);
}
