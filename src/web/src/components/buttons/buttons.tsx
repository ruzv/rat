import styles from "./buttons.module.css";

export function TextButton({
  text,
  onClick,
}: {
  text: string;
  onClick: () => void;
}) {
  return (
    <Button className={styles.textButton} onClick={onClick}>
      {text}
    </Button>
  );
}

export function IconButton({
  icon,
  onClick,
}: {
  icon: string;
  onClick: () => void;
}) {
  return (
    <Button className={styles.iconButton} onClick={onClick}>
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
  }>,
) {
  return (
    <div
      className={`${styles.button} ${props.className}`}
      onClick={props.onClick}
    >
      {props.children}
    </div>
  );
}
