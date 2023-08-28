import React from "react";
import styles from "./components.module.css";

export function Container(props: React.PropsWithChildren<{}>) {
  return <div className={styles.container}>{props.children}</div>;
}

export function Clickable(
  props: React.PropsWithChildren<{ onClick: () => void }>,
) {
  return (
    <div className={styles.clickable} onClick={props.onClick}>
      {props.children}
    </div>
  );
}

export function ClickableContainer(
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
