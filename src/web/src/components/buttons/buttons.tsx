import { InvisibleLink } from "../link";

import styles from "./buttons.module.css";

import Tooltip from "@mui/material/Tooltip";

export function TextButton({
  text,
  onClick,
  tooltip,
  href,
}: {
  text: string;
  onClick?: () => void;
  tooltip?: string;
  href?: string;
}) {
  return (
    <Button
      className={styles.textButton}
      onClick={onClick}
      tooltip={tooltip}
      href={href}
    >
      <span>{text}</span>
    </Button>
  );
}

export function IconButton({
  icon,
  onClick,
  tooltip,
  href,
}: {
  icon: string;
  onClick?: () => void;
  tooltip?: string;
  href?: string;
}) {
  return (
    <Button
      className={styles.iconButton}
      onClick={onClick}
      tooltip={tooltip}
      href={href}
    >
      <img className={styles.icon} src={icon} alt="icon" />
    </Button>
  );
}

export function ButtonRow(props: React.PropsWithChildren<{}>) {
  return <div className={styles.buttonRow}>{props.children}</div>;
}

function Button(
  props: React.PropsWithChildren<{
    onClick?: () => void;
    className?: string;
    tooltip?: string;
    href?: string;
  }>,
) {
  return (
    <>
      <WithHref href={props.href}>
        <WithTooltip tooltip={props.tooltip}>
          <div {...props} className={`${styles.button} ${props.className}`} />
        </WithTooltip>
      </WithHref>
    </>
  );
}

function WithTooltip(props: {
  children: React.ReactElement;
  tooltip?: string;
}) {
  if (!props.tooltip) {
    return props.children;
  }

  return <Tooltip title={props.tooltip}>{props.children}</Tooltip>;
}

function WithHref(props: { children: React.ReactElement; href?: string }) {
  if (!props.href) {
    return props.children;
  }

  return <InvisibleLink href={props.href}>{props.children}</InvisibleLink>;
}
