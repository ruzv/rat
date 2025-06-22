import {
  ButtonRow,
  IconTextButton,
  TextButton,
} from "./components/buttons/buttons";
import { Container } from "./components/parts";
import { Spacer } from "./components/util";

import rootIcon from "./components/icons/root.png";
import ratIcon from "./components/icons/rat.png";
import githubIcon from "./components/icons/github.png";
import signInIcon from "./components/icons/log-in.png";

export function Landing() {
  return (
    <>
      <LandingConsole />
      <Spacer height={20} />
      <Container>
        <Spacer height={30} />
        <h1>Welcome to Rat</h1>

        <p>Kinda like Hugo, Notion and Obsidian combined</p>

        <Spacer height={30} />
      </Container>
    </>
  );
}

export function LandingConsole() {
  return (
    <Container>
      <Spacer height={30} />
      <ButtonRow>
        <IconTextButton
          icon={ratIcon}
          text={"Rat"}
          tooltip="navigate to landing page"
          href={"/"}
        />
        <IconTextButton
          icon={rootIcon}
          text={"Graph Root"}
          tooltip="navigate to this graphs root node"
          href={"/view"}
        />
        <IconTextButton
          icon={signInIcon}
          text={"Sign-In"}
          tooltip="navigate to Sign-in page"
          href={"/sign-in"}
        />
      </ButtonRow>
      <Spacer height={6} />
      <ButtonRow>
        <IconTextButton
          icon={githubIcon}
          text={"GitHub"}
          tooltip="navigate to Rat projects GitHub page"
          href={"https://github.com/ruzv/rat"}
        />
        <TextButton
          text={"Credits"}
          tooltip="navigate to credits page"
          href={"/credits"}
        />
      </ButtonRow>
      <Spacer height={30} />
    </Container>
  );
}
