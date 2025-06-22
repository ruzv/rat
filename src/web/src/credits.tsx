import { Link } from "./components/link";
import { Container } from "./components/parts";
import { Spacer } from "./components/util";
import { LandingConsole } from "./landing";

export function Credits() {
  return (
    <>
      <LandingConsole />
      <Spacer height={20} />

      <Container>
        <Spacer height={30} />

        <h1>Credits</h1>

        <p>
          Created by - <Link href="https://github.com/ruzv">ruzv</Link>
        </p>

        <h2>Icons</h2>

        <ul>
          <li>
            <Link href="https://www.flaticon.com/free-icons/trash">
              Trash icons created by Freepik - Flaticon
            </Link>
          </li>

          <li>
            <Link href="https://www.flaticon.com/free-icons/search">
              Search icons created by Smashicons - Flaticon
            </Link>
          </li>

          <li>
            <Link href="https://www.flaticon.com/free-icons/new">
              New icons created by Ilham Fitrotul Hayat - Flaticon
            </Link>
          </li>

          <li>
            <Link href="https://www.flaticon.com/free-icons/tick">
              Tick icons created by Maxim Basinski Premium - Flaticon
            </Link>
          </li>

          <li>
            <Link href="https://www.flaticon.com/free-icons/close">
              Close icons created by ariefstudio - Flaticon
            </Link>
          </li>

          <li>
            <Link href="https://www.flaticon.com/free-icons/root">
              Root icons created by Khoirul Huda - Flaticon
            </Link>
          </li>

          <li>
            <Link href="https://www.flaticon.com/free-icons/copy">
              Copy icons created by Freepik - Flaticon
            </Link>
          </li>

          <li>
            <Link href="https://www.flaticon.com/free-icons/github">
              Github icons created by Pixel perfect - Flaticon
            </Link>
          </li>

          <li>
            <Link href="https://www.flaticon.com/free-icons/log-in">
              Log in icons created by Pixel perfect - Flaticon
            </Link>
          </li>
        </ul>

        <Spacer height={30} />
      </Container>
    </>
  );
}
