import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { LandingPageLegacy } from "./LandingPageLegacy";

describe("LandingPageLegacy", () => {
  it("renders Slice 11.1 static marketing content for rollback review", () => {
    render(
      <MemoryRouter>
        <LandingPageLegacy />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "Your server, under your control." })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Get Vyntrio OS" })).toHaveAttribute("href", "/download");
    expect(screen.getByRole("link", { name: "Sign in to your appliance" })).toHaveAttribute(
      "href",
      "/login"
    );
    expect(screen.getByText(/local appliance platform/i)).toBeInTheDocument();
  });
});
