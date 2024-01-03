import { signIn } from "../api/auth";

import { sessionAtom } from "../components/atoms";

import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import { Container, Spacer } from "../components/util";
import { useNavigate } from "react-router-dom";
import { useSetAtom } from "jotai";

export function SignIn() {
  const navigate = useNavigate();
  const setSession = useSetAtom(sessionAtom);

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    console.log({
      email: data.get("username"),
      password: data.get("password"),
    });

    signIn(data.get("username") as string, data.get("password") as string).then(
      (token) => {
        localStorage.setItem("rat_api_token", token);
        setSession({
          token: token,
        });
        navigate("/view");
      },
    );
  };

  return (
    <div
      style={{
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        height: "calc(100vh - 40px)",
      }}
    >
      <div
        style={{
          maxWidth: "500px",
        }}
      >
        <Box>
          <Container>
            <Spacer height={30} />
            <Typography component="h1" variant="h4">
              Sign in
            </Typography>
            <Box
              component="form"
              onSubmit={handleSubmit}
              noValidate
              sx={{ mt: 1 }}
            >
              <TextField
                margin="normal"
                required
                fullWidth
                id="username"
                label="Username"
                name="username"
                // autoComplete="username"
                autoFocus
              />
              <TextField
                margin="normal"
                required
                fullWidth
                name="password"
                label="Password"
                type="password"
                id="password"
                // autoComplete="current-password"
              />
              <Button
                type="submit"
                fullWidth
                variant="contained"
                sx={{ mt: 3, mb: 0 }}
              >
                Sign In
              </Button>
            </Box>
            <Spacer height={30} />
          </Container>
        </Box>
      </div>
    </div>
  );
}
