import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";
import { DocsPage } from "./DocsPage";

describe("DocsPage", () => {
  beforeEach(() => {
    window.localStorage.removeItem("vyntrio.locale");
  });

  it("renders German production docs without preview banner or API dependencies", () => {
    render(
      <MemoryRouter>
        <DocsPage />
      </MemoryRouter>
    );

    expect(
      screen.getByRole("heading", { name: "Dokumentation für Installation und Betrieb" })
    ).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Erste Schritte" })).toBeInTheDocument();
    expect(screen.queryByText(/Design-Vorschau/i)).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Zum Release-Bereich" })).toHaveAttribute("href", "/download");
    expect(screen.getByRole("link", { name: "Zur Startseite" })).toHaveAttribute("href", "/");
  });

  it("switches to English copy via locale switcher", () => {
    render(
      <MemoryRouter>
        <DocsPage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "EN" }));

    expect(
      screen.getByRole("heading", { name: "Documentation for install and operations" })
    ).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "First steps" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Go to release area" })).toHaveAttribute("href", "/download");
  });
});
