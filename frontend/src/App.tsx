import { AppRouter } from "./app/router";
import type { ApiClient } from "./lib/api";
import "./App.css";

type AppProps = {
  apiClient?: ApiClient;
};

export default function App({ apiClient }: AppProps) {
  return <AppRouter apiClient={apiClient} />;
}
