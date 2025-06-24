import React from "react";

import { NodeContent, ChildNodes } from "./components/parts";
import { Spacer } from "./components/util";
import { Console } from "./components/console";
import {
  nodeAtom,
  nodePathAtom,
  nodeAstAtom,
  childNodesAtom,
} from "./components/atoms";

import { useEffect, useRef } from "react";
import { useLocation } from "react-router-dom";
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
    setNode(undefined);

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
      <ScrollToAnchor />

      <Console id={node.id} />
      <Spacer height={20} />
      <NodeContent />
      <Spacer height={20} />
      <ChildNodes />
    </>
  );
}

function ScrollToAnchor() {
  const location = useLocation();
  const lastHash = useRef("");

  // listen to location change using useEffect with location as dependency
  // https://jasonwatmore.com/react-router-v6-listen-to-location-route-change-without-history-listen
  useEffect(() => {
    if (location.hash) {
      lastHash.current = location.hash.slice(1); // safe hash for further use after navigation
    }

    if (lastHash.current && document.getElementById(lastHash.current)) {
      setTimeout(() => {
        document
          .getElementById(lastHash.current)
          ?.scrollIntoView({ behavior: "smooth", block: "start" });
        lastHash.current = "";
      }, 100);
    }
  }, [location]);

  return null;
}
