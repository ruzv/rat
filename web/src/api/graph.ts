import { ratAPIBaseURL } from "./api";
import axios from "axios";

export async function search(query: string) {
  let resp = await axios.post<{ results: string[] }>(
    `${ratAPIBaseURL()}/graph/search`,
    { query },
  );

  return resp.data;
}

export async function move(id: string, newPath: string) {
  await axios.post(`${ratAPIBaseURL()}/graph/move/${id}`, { newPath });
}
