import { lazy } from "react";

export const LandingPage = lazy(() =>
  import("./landing/LandingPage").then((module) => ({ default: module.LandingPage }))
);
