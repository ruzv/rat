import { modalOpenAtom, nodePathAtom, childNodesAtom } from "./atoms";
import { TextButton, IconButton, ButtonRow } from "./buttons/buttons";
import { ConfirmModal, ContentModal } from "./modals";
import { Spacer } from "./util";
import { Code } from "./parts";
import { Node } from "./node";

import styles from "./console.module.css";
import binIcon from "./icons/bin.png";
import loupeIcon from "./icons/loupe.png";
import addNodeIcon from "./icons/add-node.png";
import { ratAPIBaseURL } from "./util";

import { useAtom, useAtomValue } from "jotai";
import { useNavigate } from "react-router-dom";
import { useState } from "react";
import { useHotkeys } from "react-hotkeys-hook";

export function Console({ id }: { id: string }) {
  const navigate = useNavigate();
  const nodePath = useAtomValue(nodePathAtom);

  if (!nodePath) {
    return <></>;
  }

  const pathParts = nodePath.split("/");

  return (
    <div className={styles.consoleContainer}>
      <ButtonRow>
        <TextButton
          text={id}
          onClick={() => {
            navigator.clipboard.writeText(id);
          }}
        />
        <TextButton
          text={nodePath}
          onClick={() => {
            navigator.clipboard.writeText(nodePath);
          }}
        />
      </ButtonRow>
      <Spacer height={6} />
      <ButtonRow>
        {pathParts.map((part, idx) => {
          return (
            <TextButton
              key={idx}
              text={part}
              onClick={() => {
                navigate(`/view/${pathParts.slice(0, idx + 1).join("/")}`);
              }}
            />
          );
        })}

        <SearchButton />
        <NewNodeButton />
        <DeleteButton pathParts={pathParts} />
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

  if (!nodePath) {
    return <></>;
  }

  return (
    <>
      <IconButton icon={addNodeIcon} onClick={showNewNodeModal} />

      {newNodeModalOpen && (
        <ContentModal>
          <Input
            handleClose={closeNewNodeModal}
            handleChange={setName}
            handleSubmit={() => {
              closeNewNodeModal();

              fetch(`${ratAPIBaseURL()}/graph/nodes/${nodePath}/`, {
                method: "POST",
                body: JSON.stringify({ name: name }),
              })
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
      <IconButton icon={loupeIcon} onClick={showSearchModal} />

      {searchModalOpen && (
        <ContentModal>
          <Input
            handleClose={closeSearchModal}
            handleChange={(query) => {
              fetch(`${ratAPIBaseURL()}/graph/search/`, {
                method: "POST",
                body: JSON.stringify({ query: query }),
              })
                .then((resp) => resp.json())
                .then((resp) => setResults(resp.results));
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
      />
      {showConfirm && (
        <ConfirmModal
          confirm={() => {
            let parentPath = pathParts.slice(0, -1).join("/");
            let path = pathParts.join("/");

            fetch(`${ratAPIBaseURL()}/graph/nodes/${path}`, {
              method: "DELETE",
            }).then((resp) => {
              if (!resp.ok) {
                console.log("failed to delete node");
                return;
              }

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