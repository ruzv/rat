import React from "react";
import { useEffect } from "react";
import { Node } from "./components/node";
import {
  Console,
  NodeContent,
  ChildNodes,
  SearchModal,
  NewNodeModal,
} from "./components/parts";
import { useAtom, useSetAtom } from "jotai";
import {
  nodeAtom,
  nodePathAtom,
  nodeAstAtom,
  childNodesAtom,
} from "./components/atoms";

function View() {
  const [node, setNode] = useAtom(nodeAtom);
  const setNodeAst = useSetAtom(nodeAstAtom);
  const setChildNodes = useSetAtom(childNodesAtom);
  const [nodePath, setNodePath] = useAtom(nodePathAtom);

  // const path = window.location.pathname.replace(/^\/view\//, "");

  // const url = "http://localhost:8889/graph/nodes/notes/";
  // console.log(path);
  // console.log(url);

  useEffect(() => {
    // if (!nodePath) {
    //   setNodePath(window.location.pathname.replace(/^\/view\//, ""));
    // }

    const url = `http://localhost:8889/graph/nodes/${nodePath}/`;

    fetch(url)
      .then((resp) => resp.json())
      .then((node: Node) => {
        setNode(node);
        setNodeAst(node.ast);
        setChildNodes(node.childNodes);
        // setNodePath(node.path);
      })
      .catch((err) => console.log(err));
  }, [nodePath, setNode, setNodeAst, setChildNodes, setNodePath]);

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