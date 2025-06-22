import React from "react";

import { Spacer } from "./util";
import { ButtonRow, IconButton } from "./buttons/buttons";

import styles from "./modals.module.css";
import confirmIcon from "./icons/confirm.png";
import cancelIcon from "./icons/cancel.png";

import { useHotkeys } from "react-hotkeys-hook";

export function ConfirmModal(
  props: React.PropsWithChildren<{
    confirm: () => void;
    cancel: () => void;
    children: React.ReactNode;
  }>,
) {
  useHotkeys("esc", props.cancel);

  return (
    <DialogModal>
      <div>{props.children}</div>
      <Spacer height={10} />
      <ButtonRow>
        <IconButton icon={confirmIcon} onClick={props.confirm} />
        <IconButton icon={cancelIcon} onClick={props.cancel} />
      </ButtonRow>
    </DialogModal>
  );
}

function DialogModal(props: React.PropsWithChildren<{}>) {
  return (
    <ModalBase>
      <div className={styles.dialogModal}>
        <ModalBox>{props.children}</ModalBox>
      </div>
    </ModalBase>
  );
}

export function ContentModal(props: React.PropsWithChildren<{}>) {
  return (
    <ModalBase>
      <div className={styles.contentModal}>
        <ModalBox>{props.children}</ModalBox>
      </div>
    </ModalBase>
  );
}

function ModalBox(props: React.PropsWithChildren<{}>) {
  return <div className={styles.modalBox}>{props.children}</div>;
}

function ModalBase(props: React.PropsWithChildren<{}>) {
  return <div className={styles.modalBase}>{props.children}</div>;
}
