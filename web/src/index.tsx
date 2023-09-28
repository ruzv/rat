import React from "react";
import ReactDOM from "react-dom/client";
import "./fonts.css";
import "./index.css";
import View from "./view";
import { createBrowserRouter, RouterProvider } from "react-router-dom";

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
    <RouterProvider router={router} />
  </React.StrictMode>,
);
