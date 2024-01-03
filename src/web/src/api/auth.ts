import { ratAPIBaseURL } from "./api";
import axios from "axios";

export async function signIn(username: string, password: string) {
  let resp = await axios.post<{ token: string }>(`${ratAPIBaseURL()}/auth`, {
    username,
    password,
  });

  return resp.data.token;
}
