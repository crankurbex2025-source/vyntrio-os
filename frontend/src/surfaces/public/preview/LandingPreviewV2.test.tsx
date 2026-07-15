import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, beforeEach } from "vitest";
import { LandingPreviewV2 } from "./LandingPreviewV2";

describe("LandingPreviewV2", () => {
  beforeEach(() => {
    window.localStorage.removeItem("vyntrio.locale");
  });

  it("renders German preview by default without API dependencies", () => {
    render(
      <MemoryRouter>
        <LandingPreviewV2 />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "Was läuft auf deinem Server?" })).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent(/Design-Vorschau/i);
    expect(screen.getAllByText(/Nach Login/i).length).toBeGreaterThan(0);
    expect(screen.getByRole("heading", { name: "Die Verwaltungsoberfläche im Überblick" })).toBeInTheDocument();
    expect(screen.queryByText("Checking session...")).not.toBeInTheDocument();
  });

  it("switches to English copy via locale switcher", () => {
    render(
      <MemoryRouter>
        <LandingPreviewV2 />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "EN" }));

    expect(screen.getByRole("heading", { name: "What runs on your server?" })).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent(/Design preview/i);
    expect(screen.getByRole("heading", { name: "What this page does not show" })).toBeInTheDocument();
  });
});
