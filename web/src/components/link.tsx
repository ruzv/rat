import { Link as RouterLink } from "react-router-dom";

import styles from "./link.module.css";

export function Link(props: React.PropsWithChildren<{ href: string }>) {
  return (
    <BaseLink href={props.href} className={styles.link}>
      {props.children}
    </BaseLink>
  );
}

export function InvisibleLink(
  props: React.PropsWithChildren<{ href: string }>,
) {
  return (
    <BaseLink href={props.href} className={styles.invisibleLink}>
      {props.children}
    </BaseLink>
  );
}

function BaseLink(
  props: React.PropsWithChildren<{ href: string; className: string }>,
) {
  // check if href is absolute, then open in new tab
  if (IsAbsoluteURL(props.href)) {
    return (
      <RouterLink
        className={props.className}
        to={props.href}
        target="_blank" // open in new tab
        rel="noopener noreferrer"
      >
        {props.children}
      </RouterLink>
    );
  }

  return (
    <RouterLink to={props.href} className={props.className}>
      {props.children}
    </RouterLink>
  );
}

const IsAbsoluteURL = (() => {
  let re = new RegExp("^(?:[a-z+]+:)?//", "i");
  return (url: string) => {
    return re.test(url);
  };
})();
