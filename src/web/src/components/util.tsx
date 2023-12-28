import styles from "./util.module.css";

import { Link as RouterLink } from "react-router-dom";

export function Spacer({
  width = 0,
  height = 0,
}: {
  width?: number;
  height?: number;
}) {
  let style = {};

  if (width !== 0) {
    style = {
      width: `${width}px`,
      height: `${width}px`,
      float: "left",
    };
  }

  if (height !== 0) {
    style = { marginBottom: `${height}px` };
  }

  return <div style={style}></div>;
}

export function InternalLink(props: React.PropsWithChildren<{ href: string }>) {
  return (
    <RouterLink to={props.href} className={styles.link}>
      {props.children}
    </RouterLink>
  );
}

export function ExternalLink(props: React.PropsWithChildren<{ href: string }>) {
  return (
    <a className={styles.link} href={props.href}>
      {props.children}
    </a>
  );
}

export function Container(props: React.PropsWithChildren<{}>) {
  return <div className={styles.container}>{props.children}</div>;
}

export function ClickableContainer(
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
