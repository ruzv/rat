import React from "react";

import { View } from "./pages/view";
import { SignIn } from "./pages/signin";
import { Landing } from "./pages/landing";

import ReactDOM from "react-dom/client";
import "@fontsource/roboto-mono/500.css";
import "./index.css";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { ThemeProvider, createTheme } from "@mui/material/styles";

const theme = createTheme({
  typography: {
    fontFamily: "Roboto Mono",
  },
  palette: {
    mode: "dark",
    text: {
      primary: "#acacac",
      secondary: "#acacac",
    },
    primary: {
      main: "#841a84",
    },
  },
});

const router = createBrowserRouter([
  {
    path: "/",
    element: <Landing />,
  },
  {
    path: "/signin",
    element: <SignIn />,
  },
  {
    path: "/view/*",
    element: <View />,
    loader: async ({ params }) => {
      return params["*"];
    },
  },
]);

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <ThemeProvider theme={theme}>
      <RouterProvider router={router} />
    </ThemeProvider>
  </React.StrictMode>,
);
