import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import type { ComponentProps } from "react";
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

  function renderShell(overrides: Partial<ComponentProps<typeof SettingsShell>> = {}) {
    const onSignOut = vi.fn();
    const onStartEdit = vi.fn();
    const onCancelEdit = vi.fn();
    const onSaveDisplayName = vi.fn();
    const onDraftDisplayNameChange = vi.fn();

    render(
      <SettingsShell
        settings={settings}
        editMode={false}
        draftDisplayName={settings.instance.name}
        isUpdating={false}
        updateErrorMessage={null}
        updateValidationMessage={null}
        onStartEdit={onStartEdit}
        onCancelEdit={onCancelEdit}
        onSaveDisplayName={onSaveDisplayName}
        onDraftDisplayNameChange={onDraftDisplayNameChange}
        isSigningOut={false}
        signOutError={false}
        onSignOut={onSignOut}
        {...overrides}
      />
    );

    return { onSignOut, onStartEdit, onCancelEdit, onSaveDisplayName, onDraftDisplayNameChange };
  }

  it("renders readonly settings values with edit and sign out actions", () => {
    renderShell();

    expect(screen.getByRole("heading", { name: "Instance settings" })).toBeInTheDocument();
    expect(screen.getByText("Vyntrio Home")).toBeInTheDocument();
    expect(screen.getByText("0.2.0-dev")).toBeInTheDocument();
    expect(screen.getByText("development")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Edit name" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Sign out" })).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Instance name")).not.toBeInTheDocument();
  });

  it("renders edit controls and callbacks while edit mode is active", () => {
    const { onSaveDisplayName, onCancelEdit, onDraftDisplayNameChange } = renderShell({
      editMode: true,
      draftDisplayName: "Draft Name",
    });

    const input = screen.getByLabelText("Instance name");
    expect(input).toHaveValue("Draft Name");
    expect(screen.getByRole("button", { name: "Save" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Edit name" })).not.toBeInTheDocument();

    fireEvent.change(input, { target: { value: "Renamed" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onDraftDisplayNameChange).toHaveBeenCalledWith("Renamed");
    expect(onSaveDisplayName).toHaveBeenCalledTimes(1);
    expect(onCancelEdit).toHaveBeenCalledTimes(1);
  });

  it("disables mutation/logout controls while update is pending", () => {
    renderShell({
      editMode: true,
      draftDisplayName: "Vyntrio Home",
      isUpdating: true,
    });

    expect(screen.getByRole("button", { name: "Saving..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Sign out" })).toBeDisabled();
    expect(screen.getByLabelText("Instance name")).toBeDisabled();
  });

  it("disables edit entry and sign out while logout is pending", () => {
    renderShell({
      isSigningOut: true,
    });

    expect(screen.getByRole("button", { name: "Edit name" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Signing out..." })).toBeDisabled();
  });

  it("renders generic update and sign-out errors only when provided", () => {
    const { rerender } = render(
      <SettingsShell
        settings={settings}
        editMode={true}
        draftDisplayName={settings.instance.name}
        isUpdating={false}
        updateErrorMessage={"The instance name could not be updated. Please try again."}
        updateValidationMessage={"Enter a valid instance name."}
        onStartEdit={vi.fn()}
        onCancelEdit={vi.fn()}
        onSaveDisplayName={vi.fn()}
        onDraftDisplayNameChange={vi.fn()}
        isSigningOut={false}
        signOutError={false}
        onSignOut={vi.fn()}
      />
    );

    expect(screen.getByText("Enter a valid instance name.")).toBeInTheDocument();
    expect(
      screen.getByText("The instance name could not be updated. Please try again.")
    ).toBeInTheDocument();
    expect(screen.queryByText("Sign-out could not be completed. Please try again.")).not.toBeInTheDocument();

    rerender(
      <SettingsShell
        settings={settings}
        editMode={false}
        draftDisplayName={settings.instance.name}
        isUpdating={false}
        updateErrorMessage={null}
        updateValidationMessage={null}
        onStartEdit={vi.fn()}
        onCancelEdit={vi.fn()}
        onSaveDisplayName={vi.fn()}
        onDraftDisplayNameChange={vi.fn()}
        isSigningOut={false}
        signOutError={true}
        onSignOut={vi.fn()}
      />
    );

    expect(
      screen.getByText("Sign-out could not be completed. Please try again.")
    ).toBeInTheDocument();
  });

  it("calls edit and sign-out actions on user interaction", () => {
    const { onSignOut, onStartEdit } = renderShell();

    fireEvent.click(screen.getByRole("button", { name: "Edit name" }));
    fireEvent.click(screen.getByRole("button", { name: "Sign out" }));

    expect(onStartEdit).toHaveBeenCalledTimes(1);
    expect(onSignOut).toHaveBeenCalledTimes(1);
  });

  it("keeps readonly rendering when not editing", () => {
    const onSignOut = vi.fn();
    render(
      <SettingsShell
        settings={settings}
        editMode={false}
        draftDisplayName={settings.instance.name}
        isUpdating={false}
        updateErrorMessage={null}
        updateValidationMessage={null}
        onStartEdit={vi.fn()}
        onCancelEdit={vi.fn()}
        onSaveDisplayName={vi.fn()}
        onDraftDisplayNameChange={vi.fn()}
        isSigningOut={false}
        signOutError={false}
        onSignOut={onSignOut}
      />
    );

    expect(screen.getByText("Vyntrio Home")).toBeInTheDocument();
    expect(screen.queryByLabelText("Instance name")).not.toBeInTheDocument();
  });
});
