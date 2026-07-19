import { useEffect, useState } from "react";
import { createApiClient } from "../../lib/api";
import { parseInstallMediaDto, type InstallMediaDto } from "./installMediaDto";

export type InstallMediaMetadataState =
  | { kind: "loading" }
  | { kind: "error" }
  | { kind: "ready"; metadata: InstallMediaDto };

export function useInstallMediaMetadata(): InstallMediaMetadataState {
  const [state, setState] = useState<InstallMediaMetadataState>({ kind: "loading" });

  useEffect(() => {
    let cancelled = false;
    const client = createApiClient();

    void (async () => {
      const result = await client.requestJson<unknown>("/api/v1/public/install-media");
      if (cancelled) {
        return;
      }
      if (!result.ok) {
        setState({ kind: "error" });
        return;
      }
      const metadata = parseInstallMediaDto(result.data);
      if (!metadata) {
        setState({ kind: "error" });
        return;
      }
      setState({ kind: "ready", metadata });
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  return state;
}
