import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";
import { DownloadPage } from "./DownloadPage";

describe("DownloadPage", () => {
  beforeEach(() => {
    window.localStorage.removeItem("vyntrio.locale");
  });

  it("renders German production download without preview banner or API dependencies", () => {
    render(
      <MemoryRouter>
        <DownloadPage />
      </MemoryRouter>
    );

    expect(
      screen.getByRole("heading", { name: "Installationsmedien und Artefakt-Status" })
    ).toBeInTheDocument();
    expect(screen.getByText("Nicht veröffentlicht")).toBeInTheDocument();
    expect(screen.queryByText(/Design-Vorschau/i)).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Zur Startseite" })).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: "Zur Dokumentation" })).toHaveAttribute("href", "/docs");
  });

  it("switches to English copy via locale switcher", () => {
    render(
      <MemoryRouter>
        <DownloadPage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "EN" }));

    expect(screen.getByRole("heading", { name: "Install media and artifact status" })).toBeInTheDocument();
    expect(screen.getByText("Not published")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Go to homepage" })).toHaveAttribute("href", "/");
  });
});
