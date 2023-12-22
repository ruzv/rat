import styles from "./buttons.module.css";

import Tooltip from "@mui/material/Tooltip";

export function TextButton({
  text,
  tooltip,
  onClick,
}: {
  text: string;
  tooltip?: string;
  onClick: () => void;
}) {
  return (
    <Button className={styles.textButton} onClick={onClick} tooltip={tooltip}>
      <span>{text}</span>
    </Button>
  );
}

export function IconButton({
  icon,
  onClick,
  tooltip,
}: {
  icon: string;
  onClick: () => void;
  tooltip?: string;
}) {
  return (
    <Button className={styles.iconButton} onClick={onClick} tooltip={tooltip}>
      <img className={styles.icon} src={icon} />
    </Button>
  );
}

export function ButtonRow(props: React.PropsWithChildren<{}>) {
  return <div className={styles.buttonRow}>{props.children}</div>;
}

function Button(
  props: React.PropsWithChildren<{
    onClick: () => void;
    className?: string;
    tooltip?: string;
  }>,
) {
  if (!props.tooltip) {
    return <div {...props} className={`${styles.button} ${props.className}`} />;
  }

  return (
    <Tooltip title={props.tooltip}>
      <div {...props} className={`${styles.button} ${props.className}`} />
    </Tooltip>
  );
}
