import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";
import { LandingPage } from "./LandingPage";

describe("LandingPage", () => {
  beforeEach(() => {
    window.localStorage.removeItem("vyntrio.locale");
  });

  it("renders German production landing without preview banner or API dependencies", () => {
    render(
      <MemoryRouter>
        <LandingPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "Was läuft auf deinem Server?" })).toBeInTheDocument();
    expect(screen.queryByText(/Design-Vorschau/i)).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Release & Download" })).toHaveAttribute("href", "/download");
    expect(screen.getByRole("link", { name: "Am Gerät anmelden" })).toHaveAttribute("href", "/login");
    expect(screen.queryByText("Checking session...")).not.toBeInTheDocument();
  });

  it("switches to English copy via locale switcher", () => {
    render(
      <MemoryRouter>
        <LandingPage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "EN" }));

    expect(screen.getByRole("heading", { name: "What runs on your server?" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Release & download" })).toHaveAttribute("href", "/download");
    expect(screen.getByRole("heading", { name: "What this page does not show" })).toBeInTheDocument();
  });
});
