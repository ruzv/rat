"use client";

import styles from "./styles.module.css";
import { FunctionComponent } from "react";
import { useState, useEffect } from "react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { darcula as SyntaxHighlighterStyle } from "react-syntax-highlighter/dist/esm/styles/prism";

interface NodeAstPart {
  type: string;
  children: NodeAstPart[];
  attributes: { [key: string]: number | string | boolean };
}

interface Node {
  id: string;
  name: string;
  path: string;
  ast: NodeAstPart;
}

export default function View({ params }: { params: { nodePath: string[] } }) {
  return <Node path={params.nodePath.join("/")} />;
}

const Node: FunctionComponent<{ path: string }> = ({ path }) => {
  const [node, setNode] = useState<Node | undefined>(undefined);

  useEffect(() => {
    fetch(`http://127.0.0.1:8889/graph/nodes/${path}/`)
      .then((resp) => resp.json())
      .then((node) => setNode(node));
  });

  if (!node) {
    return <h1>loading...</h1>;
  }

  return <>{renderAstPart(node.ast)}</>;
};

const Code: FunctionComponent<{ text: string }> = ({ text }) => {
  return <code className={styles.code}> {text} </code>;
};

const CodeBlock: FunctionComponent<{ text: string; language: string }> = ({
  text,
  language,
}) => {
  // https://github.com/react-syntax-highlighter/react-syntax-highlighter/blob/master/AVAILABLE_LANGUAGES_PRISM.MD
  if (language === "sh") {
    language = "bash";
  }

  console.log(language);

  return (
    <SyntaxHighlighter language={language} style={SyntaxHighlighterStyle}>
      {text}
    </SyntaxHighlighter>
  );
};

function renderAstPart(astPart: NodeAstPart) {
  switch (astPart.type) {
    case "document":
      if (astPart.children === undefined) {
        return <> {"empty document"} </>;
      }

      return <> {astPart.children.map(renderAstPart)} </>;
    case "heading":
      const level = astPart.attributes["level"] as number;

      switch (level) {
        case 1:
          return <h1> {astPart.children.map(renderAstPart)} </h1>;
        case 2:
          return <h2> {astPart.children.map(renderAstPart)} </h2>;
        case 3:
          return <h3> {astPart.children.map(renderAstPart)} </h3>;
        case 4:
          return <h4> {astPart.children.map(renderAstPart)} </h4>;
        case 5:
          return <h5> {astPart.children.map(renderAstPart)} </h5>;
        case 6:
          return <h6> {astPart.children.map(renderAstPart)} </h6>;
        default:
          return (
            <h1>
              {"unknown heading level"}
              {astPart.children.map(renderAstPart)}
            </h1>
          );
      }
    case "code":
      return <Code text={astPart.attributes["text"] as string} />;
    case "code_block":
      return (
        <CodeBlock
          text={astPart.attributes["text"] as string}
          language={astPart.attributes["info"] as string}
        />
      );
    case "link":
      const destination: string = astPart.attributes["destination"] as string;

      if (astPart.children === undefined) {
        <a href={destination}>{astPart.attributes["title"]}</a>;
      }

      return <a href={destination}>{astPart.children.map(renderAstPart)}</a>;
    case "list":
      if (astPart.children === undefined) {
        return <> {"empty list"} </>;
      }

      return <ul> {astPart.children.map(renderAstPart)} </ul>;
    case "list_item":
      if (astPart.children === undefined) {
        return <li> </li>;
      }

      return <li> {astPart.children.map(renderAstPart)} </li>;
    case "text":
      return <>{astPart.attributes["text"]}</>;
    case "paragraph":
      return <p> {astPart.children.map(renderAstPart)} </p>;
    case "span":
      return <span>{astPart.attributes["text"]}</span>;
    case "unknown":
      if (astPart.children === undefined) {
        return (
          <p>
            {"unknown leaf"}
            {astPart.attributes["text"]}
          </p>
        );
      }

      return (
        <p>
          {"unknown container"}
          {astPart.attributes["text"]}
          {astPart.children.map(renderAstPart)}
        </p>
      );
    default:
      if (astPart.children === undefined) {
        return (
          <p>
            {"unpimplemented parser for "}
            {astPart.type}
          </p>
        );
      }

      return (
        <p>
          {"unpimplemented parser for "}
          {astPart.type}
          {astPart.children.map(renderAstPart)}
        </p>
      );
  }
}
