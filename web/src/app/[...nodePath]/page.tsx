"use client";

import { useState, useEffect } from "react";
import { Node } from "../components/node";
import {
  Console,
  NodeContent,
  ChildNodes,
  SearchModal,
  NewNodeModal,
} from "../components/parts";
import { useAtom, useSetAtom } from "jotai";
import {
  nodeAtom,
  nodePathAtom,
  nodeAstAtom,
  childNodesAtom,
} from "../components/atoms";

export default function View({ params }: { params: { nodePath: string[] } }) {
  const [node, setNode] = useAtom(nodeAtom);
  const setNodeAst = useSetAtom(nodeAstAtom);
  const setChildNodes = useSetAtom(childNodesAtom);
  const setNodePath = useSetAtom(nodePathAtom);

  const path = params.nodePath.join("/");

  useEffect(() => {
    fetch(`${process.env.NEXT_PUBLIC_RAT_SERVER_URL}/graph/nodes/${path}/`)
      .then((resp) => resp.json())
      .then((node: Node) => {
        setNode(node);
        setNodeAst(node.ast);
        setChildNodes(node.childNodes);
        setNodePath(node.path);
      })
      .catch((err) => console.log(err));
  }, [path, setNode, setNodeAst, setChildNodes, setNodePath]);

  if (!node) {
    return <> </>;
  }

  return (
    <>
      <SearchModal />
      <NewNodeModal />
      <Console id={node.id} />
      <NodeContent />
      <ChildNodes />
    </>
  );
}
