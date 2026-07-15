import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { LandingPage } from "./LandingPage";

describe("LandingPage", () => {
  it("renders static marketing content without API dependencies", () => {
    render(
      <MemoryRouter>
        <LandingPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "Your server, under your control." })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Get Vyntrio OS" })).toHaveAttribute("href", "/download");
    expect(screen.getByRole("link", { name: "Sign in to your appliance" })).toHaveAttribute(
      "href",
      "/login"
    );
    expect(screen.getByText(/local appliance platform/i)).toBeInTheDocument();
    expect(screen.queryByText("Checking session...")).not.toBeInTheDocument();
  });
});
