import { Container } from "./components/parts";
import { Spacer } from "./components/util";
import { LandingConsole } from "./landing";

export function SignIn() {
  return (
    <>
      <LandingConsole />
      <Spacer height={20} />

      <Container>
        <Spacer height={30} />

        <h1>Sign In</h1>
        <p>NO SIGN IN FUNCTIONALITY YET</p>

        <Spacer height={30} />
      </Container>
    </>
  );
}
