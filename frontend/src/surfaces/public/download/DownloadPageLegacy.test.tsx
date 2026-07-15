import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { DownloadPageLegacy } from "./DownloadPageLegacy";

describe("DownloadPageLegacy", () => {
  it("renders Slice 11.1 download placeholder for rollback review", () => {
    render(
      <MemoryRouter>
        <DownloadPageLegacy />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "Install media coming soon" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Back to home" })).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: "Sign in to an existing appliance" })).toHaveAttribute(
      "href",
      "/login"
    );
  });
});
