import { nodeAtom, modalOpenAtom, nodePathAtom, childNodesAtom } from "./atoms";
import { TextButton, IconButton, ButtonRow } from "./buttons/buttons";
import { ConfirmModal, ContentModal } from "./modals";
import { Spacer } from "./util";
import { Code } from "./parts";
import { create, remove } from "../api/node";
import { search } from "../api/graph";

import styles from "./console.module.css";
import binIcon from "./icons/bin.png";
import loupeIcon from "./icons/loupe.png";
import addNodeIcon from "./icons/add-node.png";
import rootIcon from "./icons/root.png";

import { useAtom, useAtomValue } from "jotai";
import { useNavigate } from "react-router-dom";
import { useState } from "react";
import { useHotkeys } from "react-hotkeys-hook";

export function Console() {
  const navigate = useNavigate();
  const node = useAtomValue(nodeAtom);

  if (!node) {
    return <></>;
  }

  const isRoot = node.path === "";

  let pathParts: string[] = [];

  if (node.path) {
    pathParts = node.path.split("/");
  }

  return (
    <div className={styles.consoleContainer}>
      {!isRoot && (
        <ButtonRow>
          <TextButton
            text={node.id}
            tooltip="copy to clipboard"
            onClick={() => {
              navigator.clipboard.writeText(node.id);
            }}
          />
          <TextButton
            text={node.path}
            tooltip="copy to clipboard"
            onClick={() => {
              navigator.clipboard.writeText(node.path);
            }}
          />
        </ButtonRow>
      )}
      <Spacer height={6} />
      <ButtonRow>
        <IconButton
          icon={rootIcon}
          tooltip="navigate to root node"
          onClick={() => {
            navigate(`/view`);
          }}
        />

        {pathParts.map((part, idx) => {
          let path = pathParts.slice(0, idx + 1).join("/");

          return (
            <TextButton
              key={idx}
              text={part}
              tooltip={`navigate to ${path}`}
              onClick={() => {
                navigate(`/view/${path}`);
              }}
            />
          );
        })}

        <SearchButton />
        <NewNodeButton />
        {!isRoot && <DeleteButton pathParts={pathParts} />}
      </ButtonRow>
      <Spacer height={6} />
    </div>
  );
}

function NewNodeButton() {
  const nodePath = useAtomValue(nodePathAtom);
  const [childNodes, setChildNodes] = useAtom(childNodesAtom);
  const [modalOpen, setModalOpen] = useAtom(modalOpenAtom);
  const [newNodeModalOpen, setNewNodeModalOpen] = useState(false);
  const [name, setName] = useState("");

  function showNewNodeModal() {
    if (newNodeModalOpen) {
      closeNewNodeModal();
      return;
    }

    if (modalOpen) {
      // another modal is open
      return;
    }

    // open
    setNewNodeModalOpen(true);
    setModalOpen(true);
    setName("");
  }

  function closeNewNodeModal() {
    if (!newNodeModalOpen) {
      // not open
      return;
    }

    setNewNodeModalOpen(false);
    setModalOpen(false);
  }

  useHotkeys("ctrl+shift+k", showNewNodeModal);
  useHotkeys("meta+shift+k", showNewNodeModal);
  useHotkeys("esc", closeNewNodeModal);

  return (
    <>
      <IconButton
        icon={addNodeIcon}
        onClick={showNewNodeModal}
        tooltip="add new sub node"
      />

      {newNodeModalOpen && (
        <ContentModal>
          <Input
            handleClose={closeNewNodeModal}
            handleChange={setName}
            handleSubmit={() => {
              closeNewNodeModal();

              create(nodePath, name).then((node) => {
                if (!childNodes) {
                  setChildNodes([node]);
                  return;
                }

                setChildNodes([...childNodes, node]);
              });
            }}
          />
        </ContentModal>
      )}
    </>
  );
}

function SearchButton() {
  const navigate = useNavigate();
  const [modalOpen, setModalOpen] = useAtom(modalOpenAtom);
  const [searchModalOpen, setSearchModalOpen] = useState(false);
  const [results, setResults] = useState<string[]>([]);

  function showSearchModal() {
    if (searchModalOpen) {
      closeSearchModal();
      return;
    }

    if (modalOpen) {
      // another modal is open
      return;
    }

    setSearchModalOpen(true);
    setModalOpen(true);
  }

  function closeSearchModal() {
    if (!searchModalOpen) {
      return;
    }

    setSearchModalOpen(false);
    setModalOpen(false);
  }

  useHotkeys("ctrl+k", showSearchModal);
  useHotkeys("meta+k", showSearchModal);
  useHotkeys("esc", closeSearchModal);

  return (
    <>
      <IconButton
        icon={loupeIcon}
        onClick={showSearchModal}
        tooltip="search by path"
      />

      {searchModalOpen && (
        <ContentModal>
          <Input
            handleClose={closeSearchModal}
            handleChange={(query) => {
              search(query).then((resp) => setResults(resp.results));
            }}
            handleSubmit={() => {
              if (results.length === 0) {
                return;
              }

              closeSearchModal();
              navigate(`/view/${results[0]}`);
            }}
          />
          <SearchResults results={results} />
        </ContentModal>
      )}
    </>
  );
}

function SearchResults({ results }: { results: string[] }) {
  return (
    <div className={styles.searchResults}>
      {results.map((result, idx) => {
        return (
          <div className={styles.searchResult} key={idx}>
            {result}
          </div>
        );
      })}
    </div>
  );
}

function DeleteButton({ pathParts }: { pathParts: string[] }) {
  const navigate = useNavigate();
  const [modalOpen, setModalOpen] = useAtom(modalOpenAtom);
  const [showConfirm, setShowConfirm] = useState(false);

  return (
    <>
      <IconButton
        icon={binIcon}
        onClick={() => {
          if (modalOpen) {
            return;
          }
          setShowConfirm(true);
          setModalOpen(true);
        }}
        tooltip="delete node"
      />
      {showConfirm && (
        <ConfirmModal
          confirm={() => {
            let parentPath = pathParts.slice(0, -1).join("/");
            let path = pathParts.join("/");

            remove(path).then(() => {
              navigate(`/view/${parentPath}`);
              setModalOpen(false);
              setShowConfirm(false);
            });
          }}
          cancel={() => {
            setModalOpen(false);
            setShowConfirm(false);
          }}
        >
          <span>Are you sure you want to delete?</span>
          <Spacer height={5} />
          <NodePathParts pathParts={pathParts} />
        </ConfirmModal>
      )}
    </>
  );
}

function NodePathParts({ pathParts }: { pathParts: string[] }) {
  return (
    <div className={styles.nodePathPartsRow}>
      {pathParts.map((part, idx) => {
        return <Code text={part} key={idx} />;
      })}
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
