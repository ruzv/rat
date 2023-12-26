import React from "react";

import { NodeContent, ChildNodes } from "./components/parts";
import { Console } from "./components/console";
import {
  nodeAtom,
  nodePathAtom,
  nodeAstAtom,
  childNodesAtom,
} from "./components/atoms";

import { useEffect } from "react";
import { useAtom, useSetAtom } from "jotai";
import { useLoaderData } from "react-router-dom";

import { read } from "./api/node";

export function View() {
  const [node, setNode] = useAtom(nodeAtom);
  const setNodeAst = useSetAtom(nodeAstAtom);
  const setChildNodes = useSetAtom(childNodesAtom);
  const setNodePath = useSetAtom(nodePathAtom);

  const path = useLoaderData() as string; // path from router

  useEffect(() => {
    read(path).then((node) => {
      setNode(node);
      setNodeAst(node.ast);
      setChildNodes(node.childNodes);
      setNodePath(node.path);

      document.title = node.name;
    });
  }, [path, setNode, setNodeAst, setChildNodes, setNodePath]);

  if (!node) {
    return <> </>;
  }

  return (
    <>
      <Console id={node.id} />
      <NodeContent />
      <ChildNodes />
    </>
  );
}
