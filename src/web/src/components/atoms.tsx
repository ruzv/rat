import { atom } from "jotai";
import { Node, NodeAstPart } from "../types/node";
import { Session } from "../types/session";

export const sessionAtom = atom<Session | undefined>(
  (() => {
    let localAPIToken = localStorage.getItem("rat_api_token");

    if (!localAPIToken || localAPIToken === "") {
      return;
    }

    //TODO: parse jwt, check expiration, populate other fields.

    return {
      token: localAPIToken,
    };
  })(),
);

export const nodeAtom = atom<Node | undefined>(undefined);

export const nodeIDAtom = atom<string | undefined>(undefined);

export const nodePathAtom = atom<string>(
  window.location.pathname.replace(/^\/view\//, ""),
);

export const nodeAstAtom = atom<NodeAstPart | undefined>(undefined);

export const childNodesAtom = atom<Node[] | undefined>(undefined);

export const modalOpenAtom = atom(false);
