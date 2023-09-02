"use client";

import styles from "./styles.module.css";
import { useRouter } from "next/navigation";
import { useState, useEffect, ReactNode } from "react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { darcula as SyntaxHighlighterStyle } from "react-syntax-highlighter/dist/esm/styles/prism";

interface NodeAstPart {
  type: string;
  children: NodeAstPart[];
  attributes: { [key: string]: any };
}

interface Node {
  id: string;
  name: string;
  path: string;
  length: number;
  childNodePaths: string[];
  ast: NodeAstPart;
}

export default function View({ params }: { params: { nodePath: string[] } }) {
  return <NodeWithChildNodes pathParts={params.nodePath} />;
}

function NodeWithChildNodes({ pathParts }: { pathParts: string[] }) {
  const [node, setNode] = useState<Node | undefined>(undefined);
  const path = pathParts.join("/");

  useEffect(() => {
    fetch(`http://127.0.0.1:8889/graph/nodes/${path}/`)
      .then((resp) => resp.json())
      .then((node) => setNode(node));
  }, []);

  if (!node) {
    return <> </>;
  }

  return (
    <>
      <Console id={node.id} path={path} pathParts={pathParts} />

      <Container>
        <div className={styles.contentSpacer}> </div>
        <NodePart part={node.ast} />
        <div className={styles.contentSpacer}> </div>
      </Container>

      <ChildNodes childNodePaths={node.childNodePaths} />
    </>
  );
}

function Console({
  id,
  path,
  pathParts,
}: {
  id: string;
  path: string;
  pathParts: string[];
}) {
  const router = useRouter();

  return (
    <div className={styles.consoleContainer}>
      <div className={styles.consoleRowSpacer}>
        <ConsoleButton
          text={id}
          onClick={() => {
            navigator.clipboard.writeText(id);
          }}
        />
        <ConsoleButton
          text={path}
          onClick={() => {
            navigator.clipboard.writeText(path);
          }}
        />
      </div>
      <div>
        {pathParts.map((part, idx) => {
          return (
            <ConsoleButton
              key={idx}
              text={part}
              onClick={() => {
                router.push(`/${pathParts.slice(0, idx + 1).join("/")}/`);
              }}
            />
          );
        })}
      </div>
    </div>
  );
}

function ConsoleButton({
  text,
  onClick,
}: {
  text: string;
  onClick: () => void;
}) {
  return (
    <div
      className={`${styles.consoleButton} ${styles.clickable}`}
      onClick={onClick}
    >
      {text}
    </div>
  );
}

function Node({ path }: { path: string }) {
  const [node, setNode] = useState<Node | undefined>(undefined);

  useEffect(() => {
    fetch(`http://127.0.0.1:8889/graph/nodes/${path}/`)
      .then((resp) => resp.json())
      .then((node) => setNode(node));
  }, []);

  if (!node) {
    return <> </>;
  }

  return <NodePart part={node.ast} />;
}

function ChildNodes({ childNodePaths }: { childNodePaths: string[] }) {
  if (childNodePaths === undefined || childNodePaths.length === 0) {
    return <></>;
  }

  const [nodes, setNodes] = useState<Node[] | undefined>(undefined);

  useEffect(() => {
    const ratServer = "127.0.0.1:8889";

    Promise.all(
      childNodePaths.map(async (childNodePath) => {
        const resp = await fetch(
          `http://${ratServer}/graph/nodes/${childNodePath}/`,
        );

        return await resp.json();
      }),
    ).then((nodes) => setNodes(nodes));
  }, []);

  if (!nodes) {
    return <></>;
  }

  let leftChildNodes = [];
  let leftChildNodesLength = 0;
  let rightChildNodes = [];
  let rightChildNodesLength = 0;

  for (const n of nodes) {
    if (leftChildNodesLength > rightChildNodesLength) {
      rightChildNodes.push(n);
      rightChildNodesLength += n.length;
    } else {
      leftChildNodes.push(n);
      leftChildNodesLength += n.length;
    }
  }

  const router = useRouter();

  return (
    <div className={styles.childNodesContainer}>
      <div className={styles.childNodesColumn}>
        {leftChildNodes.map((node, idx) => (
          <ClickableContainer onClick={() => router.push(`/${node.path}/`)}>
            <div className={styles.contentSpacer}> </div>
            <NodePart key={idx} part={node.ast} />
            <div className={styles.contentSpacer}> </div>
          </ClickableContainer>
        ))}
      </div>
      <div className={styles.childNodesColumnSpacer}></div>
      <div className={styles.childNodesColumn}>
        {rightChildNodes.map((node, idx) => (
          <ClickableContainer onClick={() => router.push(`/${node.path}/`)}>
            <div className={styles.contentSpacer}> </div>
            <NodePart key={idx} part={node.ast} />
            <div className={styles.contentSpacer}> </div>
          </ClickableContainer>
        ))}
      </div>
    </div>
  );
}

function NodePart({ part }: { part: NodeAstPart }) {
  switch (part.type) {
    case "document":
      return <Document part={part} />;
    case "heading":
      return <Heading part={part} />;
    case "code":
      return <Code part={part} />;
    case "code_block":
      return <CodeBlock part={part} />;
    case "link":
      return <Link part={part} />;
    case "list":
      return <List part={part} />;
    case "list_item":
      return <ListItem part={part} />;
    case "text":
      return <>{part.attributes["text"]}</>;
    case "paragraph":
      return <Paragraph part={part} />;
    case "span":
      return <span>{part.attributes["text"]}</span>;
    case "todo":
      return <Todo part={part} />;
    case "todo_entry":
      return <TodoEntry part={part} />;
    case "html_block":
      return (
        <>
          <p>{part.attributes["text"]}</p>
        </>
      );

    case "unknown":
      if (part.children === undefined) {
        return (
          <p>
            {"unknown leaf"}
            {part.attributes["text"]}
          </p>
        );
      }

      return (
        <p>
          {"unknown container"}
          {part.attributes["text"]}
          <NodePartChildren part={part} />
        </p>
      );
    default:
      if (part.children === undefined) {
        return (
          <p>
            {"unpimplemented parser for "}
            {part.type}
          </p>
        );
      }

      return (
        <p>
          {"unpimplemented parser for "}
          {part.type}
          {part.children.map((child, idx) => (
            <NodePart key={idx} part={child} />
          ))}
        </p>
      );
  }
}

function NodePartChildren({ part }: { part: NodeAstPart }) {
  if (part.children === undefined || part.children.length === 0) {
    return <> </>;
  }

  return (
    <>
      {part.children.map((child, idx) => (
        <NodePart key={idx} part={child} />
      ))}
    </>
  );
}

function Document({ part }: { part: NodeAstPart }) {
  return <NodePartChildren part={part} />;
}

function Todo({ part }: { part: NodeAstPart }) {
  return (
    <div>
      {part.attributes["hints"].map(
        (hint: { type: string; value: any }, idx: number) => {
          return (
            <div key={idx}>
              {" "}
              {hint.type} {hint.value}{" "}
            </div>
          );
        },
      )}
      <NodePartChildren part={part} />
    </div>
  );
}

function TodoEntry({ part }: { part: NodeAstPart }) {
  return (
    <div className={styles.todoEntry}>
      <input
        className={styles.todoEntryCheckbox}
        type={"checkbox"}
        checked={part.attributes["done"] as boolean}
      />
      {part.attributes["text"]}
    </div>
  );
}

function Link({ part }: { part: NodeAstPart }) {
  if (part.children === undefined) {
    return (
      <a href={part.attributes["destination"] as string}>
        {part.attributes["title"]}
      </a>
    );
  }

  return (
    <a
      style={{ overflowWrap: "anywhere" }}
      href={part.attributes["destination"] as string}
    >
      <NodePartChildren part={part} />
    </a>
  );
}

function Heading({ part }: { part: NodeAstPart }) {
  switch (part.attributes["level"] as number) {
    case 1:
      return (
        <h1>
          <NodePartChildren part={part} />
        </h1>
      );
    case 2:
      return (
        <h2>
          <NodePartChildren part={part} />
        </h2>
      );
    case 3:
      return (
        <h3>
          <NodePartChildren part={part} />
        </h3>
      );
    case 4:
      return (
        <h4>
          <NodePartChildren part={part} />
        </h4>
      );
    case 5:
      return (
        <h5>
          <NodePartChildren part={part} />
        </h5>
      );
    case 6:
      return (
        <h6>
          <NodePartChildren part={part} />
        </h6>
      );
    default:
      return (
        <h1>
          {"unknown heading level"}
          <NodePartChildren part={part} />
        </h1>
      );
  }
}

function Code({ part }: { part: NodeAstPart }) {
  return (
    <code className={styles.code}> {part.attributes["text"] as string} </code>
  );
}

function CodeBlock({ part }: { part: NodeAstPart }) {
  let language = part.attributes["info"] as string;

  // https://github.com/react-syntax-highlighter/react-syntax-highlighter/blob/master/AVAILABLE_LANGUAGES_PRISM.MD
  if (language === "sh") {
    language = "bash";
  }

  return (
    <SyntaxHighlighter
      language={language}
      style={SyntaxHighlighterStyle}
      wrapLines={true}
      wrapLongLines={false}
      useInlineStyles={true}
    >
      {part.attributes["text"] as string}
    </SyntaxHighlighter>
  );
}

function List({ part }: { part: NodeAstPart }) {
  // {(part.attributes["ordered"] as boolean) && <p>ordered</p>}
  // {(part.attributes["definition"] as boolean) && <p>definition</p>}
  // {(part.attributes["term"] as boolean) && <p>term</p>}

  return (
    <ul>
      <NodePartChildren part={part} />
    </ul>
  );
}

function ListItem({ part }: { part: NodeAstPart }) {
  return (
    <li>
      <NodePartChildren part={part} />
    </li>
  );
}

function Paragraph({ part }: { part: NodeAstPart }) {
  return (
    <p>
      <NodePartChildren part={part} />
    </p>
  );
}

function Container(props: React.PropsWithChildren<{}>) {
  return <div className={styles.container}>{props.children}</div>;
}

function Clickable(props: React.PropsWithChildren<{ onClick: () => void }>) {
  return (
    <div className={styles.clickable} onClick={props.onClick}>
      {props.children}
    </div>
  );
}

function ClickableContainer(
  props: React.PropsWithChildren<{ onClick: () => void }>,
) {
  return (
    <div
      className={`${styles.container} ${styles.clickable}`}
      onClick={props.onClick}
    >
      {props.children}
    </div>
  );
}
