import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import App from "./App";

describe("App", () => {
  it("renders the Vyntrio OS frontend foundation content", () => {
    render(<App />);

    expect(
      screen.getByRole("heading", { name: "Vyntrio OS" })
    ).toBeInTheDocument();
    expect(screen.getByText("Dashboard foundation")).toBeInTheDocument();
    expect(
      screen.getByText("Frontend toolchain initialized")
    ).toBeInTheDocument();
  });
});
