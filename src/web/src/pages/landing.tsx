import { InternalLink } from "../components/util";

export function Landing() {
  return (
    <div>
      <h1>Landing</h1>
      <div>
        <InternalLink href="/signin">Sign In</InternalLink>
      </div>
      <div>
        <InternalLink href="/view">View</InternalLink>
      </div>
    </div>
  );
}
