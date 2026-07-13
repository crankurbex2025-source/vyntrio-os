import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { SettingsShell } from "./SettingsShell";

describe("SettingsShell", () => {
  const settings = {
    instance: {
      name: "Vyntrio Home",
      version: "0.2.0-dev",
    },
    api: {
      environment: "development",
    },
  };

  it("renders readonly settings values and sign out action", () => {
    const onSignOut = vi.fn();
    render(
      <SettingsShell
        settings={settings}
        isSigningOut={false}
        signOutError={false}
        onSignOut={onSignOut}
      />
    );

    expect(screen.getByRole("heading", { name: "Instance settings" })).toBeInTheDocument();
    expect(screen.getByText("Vyntrio Home")).toBeInTheDocument();
    expect(screen.getByText("0.2.0-dev")).toBeInTheDocument();
    expect(screen.getByText("development")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Sign out" })).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("disables sign out while pending and shows generic error only when requested", () => {
    const onSignOut = vi.fn();
    const { rerender } = render(
      <SettingsShell settings={settings} isSigningOut={true} signOutError={false} onSignOut={onSignOut} />
    );

    expect(screen.getByRole("button", { name: "Signing out..." })).toBeDisabled();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();

    rerender(
      <SettingsShell settings={settings} isSigningOut={false} signOutError={true} onSignOut={onSignOut} />
    );
    expect(
      screen.getByText("Sign-out could not be completed. Please try again.")
    ).toBeInTheDocument();
  });

  it("calls onSignOut on user interaction", () => {
    const onSignOut = vi.fn();
    render(
      <SettingsShell
        settings={settings}
        isSigningOut={false}
        signOutError={false}
        onSignOut={onSignOut}
      />
    );

    fireEvent.click(screen.getByRole("button", { name: "Sign out" }));
    expect(onSignOut).toHaveBeenCalledTimes(1);
  });
});
