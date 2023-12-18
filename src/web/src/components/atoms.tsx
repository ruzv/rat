import { atom } from "jotai";
import { Node, NodeAstPart } from "./node";

export const nodeAtom = atom<Node | undefined>(undefined);

export const nodePathAtom = atom<string>(
  window.location.pathname.replace(/^\/view\//, ""),
);

export const nodeAstAtom = atom<NodeAstPart | undefined>(undefined);

export const childNodesAtom = atom<Node[] | undefined>(undefined);

export const modalOpenAtom = atom(false);
