import { Node, NodeAstPart } from "./node";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { darcula as SyntaxHighlighterStyle } from "react-syntax-highlighter/dist/esm/styles/prism";
import { default as NextJSLink } from "next/link";
import styles from "./parts.module.css";
import { useRouter } from "next/navigation";
import { useState, useEffect, useMemo } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import { useAtom } from "jotai";
import { nodePathAtom, nodeAstAtom, childNodesAtom } from "./atoms";
import { graphviz } from "d3-graphviz";

export function Console({ id }: { id: string }) {
  const [nodePath, _] = useAtom(nodePathAtom);
  if (!nodePath) {
    return <></>;
  }

  const pathParts = nodePath.split("/");

  const router = useRouter();

  return (
    <div className={styles.consoleContainer}>
      <div>
        <ConsoleButton
          text={id}
          onClick={() => {
            navigator.clipboard.writeText(id);
          }}
        />
        <ConsoleButton
          text={nodePath}
          onClick={() => {
            navigator.clipboard.writeText(nodePath);
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

export function NodeContent() {
  const [ast, _] = useAtom(nodeAstAtom);

  if (!ast) {
    return <></>;
  }

  return (
    <NodeContainer>
      <div className={styles.contentSpacer}> </div>
      <NodePart part={ast} />
      <div className={styles.contentSpacer}> </div>
    </NodeContainer>
  );
}

export function ChildNodes() {
  const [childNodes, _] = useAtom(childNodesAtom);

  if (!childNodes || childNodes.length === 0) {
    return <></>;
  }

  return <ChildNodesColumns childNodes={childNodes} />;
}

function ChildNodesColumns({ childNodes }: { childNodes: Node[] }) {
  childNodes.sort((a, b) => {
    if (a.name < b.name) {
      return -1;
    } else if (a.name > b.name) {
      return 1;
    } else {
      return 0;
    }
  });

  // TODO: store the sorted nodes mby.

  let leftChildNodes = [];
  let leftChildNodesLength = 0;
  let rightChildNodes = [];
  let rightChildNodesLength = 0;

  for (const n of childNodes) {
    if (leftChildNodesLength > rightChildNodesLength) {
      rightChildNodes.push(n);
      rightChildNodesLength += n.length;
    } else {
      leftChildNodes.push(n);
      leftChildNodesLength += n.length;
    }
  }

  return (
    <div className={styles.childNodesContainer}>
      <ChildNodesColumn childNodes={leftChildNodes} />
      <div className={styles.childNodesColumnSpacer}> </div>
      <ChildNodesColumn childNodes={rightChildNodes} />
    </div>
  );
}

function ChildNodesColumn({ childNodes }: { childNodes: Node[] }) {
  return (
    <div className={styles.childNodesColumn}>
      {childNodes.map((node) => (
        <ChildNode key={node.id} node={node} />
      ))}
    </div>
  );
}

function ChildNode({ node }: { node: Node }) {
  const router = useRouter();

  return (
    <ChildNodeContainer onClick={() => router.push(`/${node.path}/`)}>
      <div className={styles.contentSpacer}></div>
      <NodePart part={node.ast} />
      <div className={styles.contentSpacer}></div>
    </ChildNodeContainer>
  );
}

export function NodePart({ part }: { part: NodeAstPart }) {
  switch (part.type) {
    case "document":
      return <Document part={part} />;
    case "heading":
      return <Heading part={part} />;
    case "horizontal_rule":
      return <hr className={styles.horizontalRule} />;
    case "code":
      return <Code part={part} />;
    case "code_block":
      return <CodeBlock part={part} />;
    case "link":
      return <Link part={part} />;
    case "graph_link":
      return <GraphLink part={part} />;
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
    case "kanban":
      return <Kanban part={part} />;
    case "kanban_column":
      return <KanbanColumn part={part} />;
    case "kanban_card":
      return <KanbanCard part={part} />;
    case "graphviz":
      return <Graphviz dot={part.attributes["text"]} />;
    case "table":
      return (
        <table className={styles.table}>
          <NodePartChildren part={part} />
        </table>
      );
    case "table_header":
      return <NodePartChildren part={part} />;
    case "table_body":
      return <NodePartChildren part={part} />;

    case "table_row":
      return (
        <tr className={styles.tableRow}>
          <NodePartChildren part={part} />
        </tr>
      );
    case "table_cell":
      return (
        <td className={styles.tableData}>
          <NodePartChildren part={part} />
        </td>
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
            {"unimplemented parser for "}
            {part.type}
          </p>
        );
      }

      return (
        <p>
          {"unimplemented parser for "}
          {part.type}
          {part.children.map((child, idx) => (
            <NodePart key={idx} part={child} />
          ))}
        </p>
      );
  }
}

function Document({ part }: { part: NodeAstPart }) {
  return <NodePartChildren part={part} />;
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
    <code className={styles.code}>{part.attributes["text"] as string}</code>
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
      customStyle={{ borderRadius: "8px" }}
    >
      {part.attributes["text"] as string}
    </SyntaxHighlighter>
  );
}

function Link({ part }: { part: NodeAstPart }) {
  if (part.children === undefined) {
    return (
      <a
        className={styles.link}
        href={part.attributes["destination"] as string}
      >
        {part.attributes["title"]}
      </a>
    );
  }

  return (
    <a className={styles.link} href={part.attributes["destination"] as string}>
      <NodePartChildren part={part} />
    </a>
  );
}

function GraphLink({ part }: { part: NodeAstPart }) {
  return (
    <NextJSLink
      className={styles.link}
      href={part.attributes["destination"] as string}
    >
      {part.attributes["title"]}
    </NextJSLink>
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

function Todo({ part }: { part: NodeAstPart }) {
  return (
    <div>
      <TodoHints part={part} />
      <NodePartChildren part={part} />
    </div>
  );
}

function TodoHints({ part }: { part: NodeAstPart }) {
  if (!part.attributes["hints"]) {
    return <></>;
  }

  return (
    <>
      {part.attributes["hints"].map(
        (hint: { type: string; value: any }, idx: number) => {
          return (
            <div key={idx}>
              {hint.type} {hint.value}
            </div>
          );
        },
      )}
    </>
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

function Kanban({ part }: { part: NodeAstPart }) {
  return (
    <div
      className={styles.kanban}
      style={{
        gridTemplateColumns: `repeat(${part.children.length}, minmax(0, 1fr))`,
      }}
    >
      <NodePartChildren part={part} />
    </div>
  );
}

function KanbanColumn({ part }: { part: NodeAstPart }) {
  return (
    <div>
      <h1 className={styles.kanbanColumnTitle}>{part.attributes["title"]}</h1>
      <NodePartChildren part={part} />
    </div>
  );
}

function KanbanCard({ part }: { part: NodeAstPart }) {
  return (
    <KanbanCardContainer onClick={() => {}}>
      <NodePartChildren part={part} />
    </KanbanCardContainer>
  );
}

let graphvizIDCounter = 0;
const graphvizID = () => `graphviz${graphvizIDCounter++}`;

function Graphviz({ dot }: { dot: string }) {
  const id = useMemo(graphvizID, []);

  useEffect(() => {
    try {
      graphviz(`#${id}`, {
        fit: true,
        // height: 500,
        // width: 500,
        zoom: false,
      }).renderDot(dot);
    } catch (error) {
      console.error(error);
    }
  }, [dot]);

  return <div className={styles.graphviz} id={id} />;
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

function NodeContainer(props: React.PropsWithChildren<{}>) {
  return (
    <Container className={styles.nodeContainerSpacer}>
      {props.children}
    </Container>
  );
}

function ChildNodeContainer(
  props: React.PropsWithChildren<{
    onClick: () => void | undefined;
  }>,
) {
  return (
    <Container
      className={`${styles.nodeContainerSpacer} ${styles.clickable}`}
      onClick={props.onClick}
    >
      {props.children}
    </Container>
  );
}

function KanbanCardContainer(
  props: React.PropsWithChildren<{
    onClick: () => void | undefined;
  }>,
) {
  return (
    <Container
      className={styles.kanbanCardContainerSpacer}
      onClick={props.onClick}
    >
      {props.children}
    </Container>
  );
}

function Container({
  className = "",
  onClick = () => {},
  children,
}: React.PropsWithChildren<{
  className: string | undefined;
  onClick?: () => void;
}>) {
  className += ` ${styles.container}`;

  return (
    <div className={className} onClick={onClick}>
      {children}
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

export function NewNodeModal() {
  const [nodePath, _] = useAtom(nodePathAtom);
  const [childNodes, setChildNodes] = useAtom(childNodesAtom);

  const [show, setShow] = useState(false);
  const [name, setName] = useState("");

  if (!nodePath) {
    return <></>;
  }

  useHotkeys(
    "ctrl+shift+k",
    () => {
      setName("");
      setShow(!show);
    },
    [show],
  );
  useHotkeys(
    "meta+shift+k",
    () => {
      setName("");
      setShow(!show);
    },
    [show],
  );
  useHotkeys(
    "esc",
    () => {
      setShow(false);
      setName("");
    },
    [show],
  );

  return (
    <>
      {show && (
        <Modal>
          <Input
            handleClose={() => {
              setShow(false);
              setName("");
            }}
            handleChange={setName}
            handleSubmit={() => {
              setShow(false);
              setName("");

              fetch(
                `${process.env.NEXT_PUBLIC_RAT_SERVER_URL}/graph/nodes/${nodePath}/`,
                {
                  method: "POST",
                  body: JSON.stringify({ name: name }),
                },
              )
                .then((resp) => resp.json())
                .then((node: Node) => {
                  if (!childNodes) {
                    setChildNodes([node]);
                    return;
                  }

                  setChildNodes([...childNodes, node]);
                });
            }}
          />
        </Modal>
      )}
    </>
  );
}

export function SearchModal() {
  const [show, setShow] = useState(false);
  const [query, setQuery] = useState("");
  const [submit, setSubmit] = useState(false);

  useHotkeys(
    "ctrl+k",
    () => {
      if (show) {
        setQuery("");
      }
      setShow(!show);
    },
    [show],
  );
  useHotkeys(
    "meta+k",
    () => {
      if (show) {
        setQuery("");
      }
      setShow(!show);
    },
    [show],
  );
  useHotkeys(
    "esc",
    () => {
      setShow(false);
      setQuery("");
    },
    [show],
  );

  return (
    <>
      {show && (
        <Modal>
          <Input
            handleClose={() => {
              setShow(false);
              setQuery("");
            }}
            handleChange={setQuery}
            handleSubmit={() => {
              setSubmit(true);
            }}
          />
          <SearchResults query={query} submit={submit} />
        </Modal>
      )}
    </>
  );
}

function SearchResults({ query, submit }: { query: string; submit: boolean }) {
  if (query === "") {
    return <></>;
  }

  interface Response {
    results: string[];
  }

  const [response, setResponse] = useState<Response | undefined>(undefined);

  useEffect(() => {
    fetch(`${process.env.NEXT_PUBLIC_RAT_SERVER_URL}/graph/search/`, {
      method: "POST",
      body: JSON.stringify({ query: query }),
    })
      .then((resp) => resp.json())
      .then((resp) => setResponse(resp));
  }, [query]);

  if (!response || response.results.length === 0) {
    return <></>;
  }

  if (submit) {
    const router = useRouter();
    router.push(`/${response.results[0]}/`);
  }

  return (
    <div className={styles.searchResults}>
      {response.results.map((result, idx) => {
        return (
          <div className={styles.searchResult} key={idx}>
            {result}
          </div>
        );
      })}
    </div>
  );
}

function Modal(props: React.PropsWithChildren<{}>) {
  return (
    <div className={styles.modal}>
      <div className={styles.modalMargins}>{props.children}</div>
    </div>
  );
}

function Input({
  handleClose,
  handleChange,
  handleSubmit,
}: {
  handleClose: () => void;
  handleChange: (value: string) => void;
  handleSubmit: () => void;
}) {
  function handleKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key === "Escape") {
      handleClose();
    }
  }

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        handleSubmit();
      }}
    >
      <input
        className={styles.input}
        type={"text"}
        autoFocus
        onKeyDown={handleKeyDown}
        onChange={(event) => {
          handleChange(event.target.value);
        }}
      />
    </form>
  );
}
