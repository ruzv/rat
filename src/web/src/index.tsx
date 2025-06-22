import React from "react";
import ReactDOM from "react-dom/client";
import "@fontsource/roboto-mono/500.css";
import "./index.css";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { Landing } from "./landing";
import { View } from "./view";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import { Credits } from "./credits";
import { SignIn } from "./signin";

const theme = createTheme({
  typography: {
    fontFamily: "Roboto Mono",
  },
});

const router = createBrowserRouter([
  {
    path: "/",
    element: <Landing />,
  },
  {
    path: "/credits",
    element: <Credits />,
  },
  {
    path: "/sign-in",
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
