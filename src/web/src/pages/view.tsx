import React from "react";

import { NodeContent, ChildNodes } from "../components/parts";
import { Console } from "../components/console";
import {
  sessionAtom,
  nodeAtom,
  nodePathAtom,
  nodeAstAtom,
  childNodesAtom,
} from "../components/atoms";
import { read } from "../api/node";

import { useEffect } from "react";
import { useAtom, useSetAtom, useAtomValue } from "jotai";
import { useLoaderData, useNavigate } from "react-router-dom";

export function View() {
  const navigate = useNavigate();

  const session = useAtomValue(sessionAtom);
  const [node, setNode] = useAtom(nodeAtom);
  const setNodeAst = useSetAtom(nodeAstAtom);
  const setChildNodes = useSetAtom(childNodesAtom);
  const setNodePath = useSetAtom(nodePathAtom);

  const path = useLoaderData() as string; // path from router

  const [error, setError] = React.useState<string | undefined>(undefined);

  useEffect(() => {
    if (!session) {
      navigate("/signin");

      return;
    }

    read(session.token, path)
      .then((node) => {
        setNode(node);
        setNodeAst(node.ast);
        setChildNodes(node.childNodes);
        setNodePath(node.path);

        document.title = node.name;
      })
      .catch((err) => {
        if (err.response) {
          setError(err.response.data.error);

          return;
        }

        setError(err.message);
      });
  }, [
    session,
    path,
    setNode,
    setNodeAst,
    setChildNodes,
    setNodePath,
    navigate,
  ]);

  if (error) {
    return <>{error}</>;
  }

  if (!node) {
    return <> </>;
  }

  return (
    <>
      <Console />
      <NodeContent />
      <ChildNodes />
    </>
  );
}
