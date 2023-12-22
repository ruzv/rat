import React from "react";
import ReactDOM from "react-dom/client";
import "@fontsource/roboto-mono/500.css";
import "./index.css";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { View } from "./view";
import { ThemeProvider, createTheme } from "@mui/material/styles";

const theme = createTheme({
  typography: {
    fontFamily: "Roboto Mono",
  },
});

const router = createBrowserRouter([
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
