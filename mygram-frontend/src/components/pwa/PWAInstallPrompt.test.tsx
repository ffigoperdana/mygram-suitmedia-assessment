import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";

import { PWAInstallPrompt } from "@/components/pwa/PWAInstallPrompt";

describe("PWAInstallPrompt", () => {
  it("shows install prompt only after browser install criteria event", async () => {
    render(<PWAInstallPrompt />);

    expect(screen.queryByText("Install MyGram")).not.toBeInTheDocument();

    window.dispatchEvent(
      new Event("beforeinstallprompt", {
        cancelable: true,
      }),
    );

    expect(await screen.findByText("Install MyGram")).toBeInTheDocument();
  });

  it("does not show again after dismissal", async () => {
    const { unmount } = render(<PWAInstallPrompt />);
    window.dispatchEvent(new Event("beforeinstallprompt", { cancelable: true }));

    await userEvent.click(await screen.findByRole("button", { name: "Not now" }));

    expect(localStorage.getItem("mygram-pwa-install-dismissed")).toBe("true");

    unmount();
    render(<PWAInstallPrompt />);
    window.dispatchEvent(new Event("beforeinstallprompt", { cancelable: true }));

    expect(screen.queryByText("Install MyGram")).not.toBeInTheDocument();
  });
});
