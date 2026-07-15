import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";
import { DownloadPreviewV2 } from "./DownloadPreviewV2";

describe("DownloadPreviewV2", () => {
  beforeEach(() => {
    window.localStorage.removeItem("vyntrio.locale");
  });

  it("renders German download preview without API dependencies", () => {
    render(
      <MemoryRouter>
        <DownloadPreviewV2 />
      </MemoryRouter>
    );

    expect(
      screen.getByRole("heading", { name: "Installationsmedien und Artefakt-Status" })
    ).toBeInTheDocument();
    expect(screen.getByText("Nicht veröffentlicht")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Zur Vorschau-Startseite" })).toHaveAttribute(
      "href",
      "/design-preview/landing"
    );
  });

  it("switches to English copy via locale switcher", () => {
    render(
      <MemoryRouter>
        <DownloadPreviewV2 />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "EN" }));

    expect(screen.getByRole("heading", { name: "Install media and artifact status" })).toBeInTheDocument();
    expect(screen.getByText("Not published")).toBeInTheDocument();
  });
});
