import React from "react";

import { Node } from "./components/node";
import { NodeContent, ChildNodes } from "./components/parts";
import { Console } from "./components/console";
import { ratAPIBaseURL } from "./components/util";
import {
  nodeAtom,
  nodePathAtom,
  nodeAstAtom,
  childNodesAtom,
} from "./components/atoms";

import { useEffect } from "react";
import { useAtom, useSetAtom } from "jotai";
import { useLoaderData } from "react-router-dom";

export function View() {
  const [node, setNode] = useAtom(nodeAtom);
  const setNodeAst = useSetAtom(nodeAstAtom);
  const setChildNodes = useSetAtom(childNodesAtom);
  const setNodePath = useSetAtom(nodePathAtom);

  const path = useLoaderData(); // path from router

  useEffect(() => {
    fetch(`${ratAPIBaseURL()}/graph/node/${path}`)
      .then((resp) => resp.json())
      .then((node: Node) => {
        setNode(node);
        setNodeAst(node.ast);
        setChildNodes(node.childNodes);
        setNodePath(node.path);

        document.title = node.name;
      })
      .catch((err) => console.log(err));
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
