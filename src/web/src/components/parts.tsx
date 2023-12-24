import React from "react";

import { Node, NodeAstPart } from "./node";
import { nodeAstAtom, childNodesAtom } from "./atoms";
import { Spacer, ratAPIBaseURL } from "./util";
import styles from "./parts.module.css";

import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { darcula as SyntaxHighlighterStyle } from "react-syntax-highlighter/dist/esm/styles/prism";
import { useState, useEffect, useMemo } from "react";
import { useAtomValue } from "jotai";
import { graphviz } from "d3-graphviz";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import {
  useDroppable,
  useDraggable,
  DndContext,
  DragEndEvent,
} from "@dnd-kit/core";
import _ from "lodash";
import Checkbox from "@mui/material/Checkbox";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import MuiTableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";

import { CSS } from "@dnd-kit/utilities";

function Droppable({
  id,
  children,
}: {
  id: string;
  children: React.ReactNode;
}) {
  const { setNodeRef } = useDroppable({ id: id });

  return <div ref={setNodeRef}>{children}</div>;
}

function Draggable({
  id,
  children,
}: {
  id: string;
  children: React.ReactNode;
}) {
  const { attributes, listeners, setNodeRef, transform } = useDraggable({
    id: id,
  });
  const style = {
    transform: CSS.Translate.toString(transform),
  };

  return (
    <div ref={setNodeRef} style={style} {...listeners} {...attributes}>
      {children}
    </div>
  );
}

export function NodeContent() {
  const ast = useAtomValue(nodeAstAtom);

  if (!ast) {
    return <></>;
  }

  return (
    <Container>
      <Spacer height={30} />
      <NodePart part={ast} />
      <Spacer height={30} />
    </Container>
  );
}

export function ChildNodes() {
  const childNodes = useAtomValue(childNodesAtom);

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

  let leftChildNodes: Node[] = [];
  let leftChildNodesLength = 0;
  let rightChildNodes: Node[] = [];
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
      <Spacer width={20} />
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
  const navigate = useNavigate();

  return (
    <>
      <Spacer height={20} />
      <ClickableContainer onClick={() => navigate(`/view/${node.path}`)}>
        <Spacer height={30} />
        <NodePart part={node.ast} />
        <Spacer height={30} />
      </ClickableContainer>
    </>
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
      return <Code text={part.attributes["text"] as string} />;
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
    case "image":
      return <Image part={part} />;
    case "embed":
      return <Embed url={part.attributes["url"] as string} />;
    case "rat_error":
      return <Error err={part.attributes["err"] as string} />;
    case "table":
      return (
        <MuiTableContainer component={TableContainer}>
          <Table
            sx={{ color: "#acacac" }}
            size="small"
            aria-label="a dense table"
          >
            <NodePartChildren part={part} />
          </Table>
        </MuiTableContainer>
      );
    case "table_header":
      return (
        <TableHead>
          <NodePartChildren part={part} />
        </TableHead>
      );
    case "table_body":
      return (
        <TableBody>
          <NodePartChildren part={part} />
        </TableBody>
      );
    case "table_row":
      return (
        <TableRow>
          <NodePartChildren part={part} />
        </TableRow>
      );
    case "table_cell":
      return (
        <TableCell sx={{ color: "#acacac" }}>
          <NodePartChildren part={part} />
        </TableCell>
      );
    case "strong":
      return (
        <strong>
          <NodePartChildren part={part} />
        </strong>
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

function TableContainer({ children }: { children: React.ReactNode }) {
  return <div className={styles.tableContainer}>{children}</div>;
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

export function Code({ text }: { text: string }) {
  return <code className={styles.code}>{text}</code>;
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
  const href = part.attributes["destination"] as string;

  if (part.children === undefined) {
    return (
      <a className={styles.link} href={href}>
        {part.attributes["title"]}
      </a>
    );
  }

  return (
    <a className={styles.link} href={href}>
      <NodePartChildren part={part} />
    </a>
  );
}

function GraphLink({ part }: { part: NodeAstPart }) {
  return (
    <RouterLink
      to={part.attributes["destination"] as string}
      className={styles.link}
    >
      {part.attributes["title"]}
    </RouterLink>
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
    <div className={styles.paragraph}>
      <NodePartChildren part={part} />
    </div>
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
      <div className={styles.todoEntryCheckbox}>
        <TodoCheckbox done={part.attributes["done"] as boolean} />
      </div>
      <div className={styles.todoEntryTextContainer}>
        <div className={styles.todoEntryText}>
          <NodePartChildren part={part} />
        </div>
      </div>
    </div>
  );
}

function TodoCheckbox({ done }: { done: boolean }) {
  if (!done) {
    return (
      <div className={styles.todoEntryCheckbox}>
        <Checkbox
          sx={{ "& .MuiSvgIcon-root": { fontSize: 32 } }}
          color="secondary"
        />
      </div>
    );
  }

  return (
    <div className={styles.todoEntryCheckbox}>
      <Checkbox
        sx={{ "& .MuiSvgIcon-root": { fontSize: 32 } }}
        defaultChecked
        color="secondary"
      />
    </div>
  );
}

function Kanban({ part }: { part: NodeAstPart }) {
  const [kanbanPart, setKanbanPart] = useState(part);

  if (!part.children || part.children.length === 0) {
    return <></>;
  }

  return (
    <div
      className={styles.kanban}
      style={{
        gridTemplateColumns: `repeat(${part.children.length}, minmax(0, 1fr))`,
      }}
    >
      <DndContext onDragEnd={handleDragEnd}>
        <NodePartChildren part={kanbanPart} />
      </DndContext>
    </div>
  );

  function handleDragEnd(event: DragEndEvent) {
    if (!event.over) {
      return;
    }
    let cardID = event.active.id;
    let columnID = event.over.id;
    let targetCardIdx = -1;
    let newKanbanPart = _.cloneDeep(kanbanPart);
    let srcColumn, destColumn, targetCard: NodeAstPart | undefined;

    for (
      let columnIdx = 0;
      columnIdx < newKanbanPart.children.length;
      columnIdx++
    ) {
      if (newKanbanPart.children[columnIdx].attributes["id"] !== columnID) {
        continue;
      }

      destColumn = newKanbanPart.children[columnIdx];
    }

    for (
      let columnIdx = 0;
      columnIdx < newKanbanPart.children.length;
      columnIdx++
    ) {
      const column = newKanbanPart.children[columnIdx];

      if (!column.children) {
        column.children = [];
      }

      for (let cardIdx = 0; cardIdx < column.children.length; cardIdx++) {
        const card = column.children[cardIdx];

        if (card.attributes["id"] !== cardID) {
          continue;
        }

        srcColumn = column;
        targetCard = column.children[cardIdx];
        targetCardIdx = cardIdx;
      }
    }

    if (!destColumn || !srcColumn || !targetCard) {
      return;
    }

    if (srcColumn === destColumn) {
      return;
    }

    destColumn.children.push(targetCard);
    srcColumn.children.splice(targetCardIdx, 1);

    fetch(`${ratAPIBaseURL()}/graph/move/${targetCard.attributes["id"]}`, {
      method: "POST",
      body: JSON.stringify({
        newPath: `${destColumn.attributes["path"]}/${targetCard.attributes["name"]}`,
      }),
    });

    setKanbanPart(newKanbanPart);
    console.log("done");
  }
}

function KanbanColumn({ part }: { part: NodeAstPart }) {
  return (
    <Droppable id={part.attributes["id"]}>
      <div>
        <h1 className={styles.kanbanColumnTitle}>{part.attributes["name"]}</h1>
        <NodePartChildren part={part} />
      </div>
    </Droppable>
  );
}

function KanbanCard({ part }: { part: NodeAstPart }) {
  return (
    <Draggable id={part.attributes["id"]}>
      <Spacer height={10} />
      <Container>
        <NodePartChildren part={part} />
      </Container>
    </Draggable>
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
        zoom: false,
      }).renderDot(dot);
    } catch (error) {
      console.error(error);
    }
  }, [dot, id]);

  return <div className={styles.graphviz} id={id} />;
}

function Image({ part }: { part: NodeAstPart }) {
  return (
    <img
      className={styles.image}
      src={part.attributes["src"]}
      alt={part.children[0].attributes["text"]}
    />
  );
}

function Embed({ url }: { url: string }) {
  return (
    <div className={styles.embedContainer}>
      <iframe className={styles.embedIframe} src={url} title={url} />
    </div>
  );
}

function Error({ err }: { err: string }) {
  return (
    <div className={styles.errorContainer}>
      <div className={styles.errorHeader}>ERROR</div>
      <div className={styles.errorText}>{err}</div>
    </div>
  );
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

function ClickableContainer(
  props: React.PropsWithChildren<{ onClick?: () => void }>,
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

function Container(props: React.PropsWithChildren<{}>) {
  return <div className={styles.container}>{props.children}</div>;
}
