"use client";

import { useState, useEffect } from "react";
import { Node } from "../components/node";
import {
  Console,
  NodeContent,
  ChildNodes,
  SearchModal,
} from "../components/parts";

const ratServer = "http://127.0.0.1:8889";

export default function View({ params }: { params: { nodePath: string[] } }) {
  const [node, setNode] = useState<Node | undefined>(undefined);
  const path = params.nodePath.join("/");

  useEffect(() => {
    fetch(`${ratServer}/graph/nodes/${path}/`)
      .then((resp) => resp.json())
      .then((node) => setNode(node));
  }, []);

  if (!node) {
    return <> </>;
  }

  return (
    <>
      <SearchModal />
      <Console id={node.id} path={path} pathParts={params.nodePath} />
      <NodeContent node={node} />
      <ChildNodes childNodes={node.childNodes} />
    </>
  );
}
