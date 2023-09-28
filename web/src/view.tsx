import React from "react";
import { useEffect } from "react";
import { Node } from "./components/node";
import {
  Console,
  NodeContent,
  ChildNodes,
  SearchModal,
  NewNodeModal,
  ratAPIBaseURL,
} from "./components/parts";
import { useAtom, useSetAtom } from "jotai";
import {
  nodeAtom,
  nodePathAtom,
  nodeAstAtom,
  childNodesAtom,
} from "./components/atoms";
import { useLoaderData } from "react-router-dom";

function View() {
  const [node, setNode] = useAtom(nodeAtom);
  const setNodeAst = useSetAtom(nodeAstAtom);
  const setChildNodes = useSetAtom(childNodesAtom);
  const setNodePath = useSetAtom(nodePathAtom);

  const path = useLoaderData(); // path from router

  useEffect(() => {
    fetch(`${ratAPIBaseURL()}/graph/nodes/${path}/`)
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

export default View;
