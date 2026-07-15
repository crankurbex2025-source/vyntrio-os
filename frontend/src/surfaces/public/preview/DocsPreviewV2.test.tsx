import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";
import { DocsPreviewV2 } from "./DocsPreviewV2";

describe("DocsPreviewV2", () => {
  beforeEach(() => {
    window.localStorage.removeItem("vyntrio.locale");
  });

  it("renders German docs preview without API dependencies", () => {
    render(
      <MemoryRouter>
        <DocsPreviewV2 />
      </MemoryRouter>
    );

    expect(
      screen.getByRole("heading", { name: "Dokumentation für Installation und Betrieb" })
    ).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Erste Schritte" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Appliance im Alltag" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Zur Release-Vorschau" })).toHaveAttribute(
      "href",
      "/design-preview/download"
    );
  });

  it("switches to English copy via locale switcher", () => {
    render(
      <MemoryRouter>
        <DocsPreviewV2 />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "EN" }));

    expect(
      screen.getByRole("heading", { name: "Documentation for install and operations" })
    ).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "First steps" })).toBeInTheDocument();
  });
});
