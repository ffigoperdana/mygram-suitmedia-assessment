import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach, beforeAll, vi } from "vitest";

vi.mock("cap-widget", () => {
  class MockCap {
    widget = document.createElement("cap-widget");

    addEventListener(type: string, listener: EventListener) {
      this.widget.addEventListener(type, listener);
    }

    reset() {
      this.widget.dispatchEvent(new CustomEvent("reset", { detail: {} }));
    }

    async solve() {
      this.widget.dispatchEvent(
        new CustomEvent("progress", { detail: { progress: 100 } }),
      );
      const token = "cap-token-123";
      this.widget.dispatchEvent(new CustomEvent("solve", { detail: { token } }));
      return { success: true, token };
    }
  }

  if (!customElements.get("cap-widget")) {
    customElements.define(
      "cap-widget",
      class MockCapWidget extends HTMLElement {
        reset() {
          this.dispatchEvent(new CustomEvent("reset", { detail: {} }));
        }
      },
    );
  }

  return { default: MockCap, Cap: MockCap };
});

beforeAll(() => {
  Object.defineProperty(window, "matchMedia", {
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  });

  Object.defineProperty(navigator, "clipboard", {
    configurable: true,
    value: {
      writeText: vi.fn().mockResolvedValue(undefined),
    },
  });

  Object.defineProperty(window, "confirm", {
    configurable: true,
    value: vi.fn(() => true),
  });
});

afterEach(() => {
  cleanup();
  localStorage.clear();
  sessionStorage.clear();
  vi.clearAllMocks();
  vi.unstubAllEnvs();
});
