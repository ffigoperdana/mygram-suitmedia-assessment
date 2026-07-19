import { useEffect, useMemo, useState } from "react";
import { Download, X } from "lucide-react";

import { Button } from "@/components/ui/button";

type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed"; platform: string }>;
};

const dismissedKey = "mygram-pwa-install-dismissed";

function isStandaloneDisplay() {
  return (
    window.matchMedia("(display-mode: standalone)").matches ||
    window.navigator.standalone === true
  );
}

function isIosSafari() {
  const userAgent = window.navigator.userAgent.toLowerCase();
  return /iphone|ipad|ipod/.test(userAgent) && !/crios|fxios|edgios/.test(userAgent);
}

export function PWAInstallPrompt() {
  const [deferredPrompt, setDeferredPrompt] = useState<BeforeInstallPromptEvent | null>(null);
  const [visible, setVisible] = useState(false);
  const iosSafari = useMemo(isIosSafari, []);

  useEffect(() => {
    if (localStorage.getItem(dismissedKey) === "true" || isStandaloneDisplay()) {
      return undefined;
    }

    if (iosSafari) {
      const timer = window.setTimeout(() => setVisible(true), 1500);
      return () => window.clearTimeout(timer);
    }

    function handleBeforeInstallPrompt(event: Event) {
      event.preventDefault();
      setDeferredPrompt(event as BeforeInstallPromptEvent);
      setVisible(true);
    }

    window.addEventListener("beforeinstallprompt", handleBeforeInstallPrompt);
    return () => window.removeEventListener("beforeinstallprompt", handleBeforeInstallPrompt);
  }, [iosSafari]);

  async function handleInstall() {
    if (!deferredPrompt) {
      dismiss();
      return;
    }

    await deferredPrompt.prompt();
    await deferredPrompt.userChoice;
    dismiss();
  }

  function dismiss() {
    localStorage.setItem(dismissedKey, "true");
    setVisible(false);
    setDeferredPrompt(null);
  }

  if (!visible) {
    return null;
  }

  return (
    <div className="fixed inset-x-3 bottom-3 z-50 mx-auto max-w-md rounded-lg border bg-card p-4 shadow-lg">
      <div className="flex items-start gap-3">
        <span className="grid h-10 w-10 shrink-0 place-items-center rounded-md bg-primary text-primary-foreground">
          <Download className="h-5 w-5" aria-hidden="true" />
        </span>
        <div className="min-w-0 flex-1">
          <p className="font-medium">Install MyGram</p>
          <p className="mt-1 text-sm text-muted-foreground">
            {iosSafari
              ? "Use Share, then Add to Home Screen to install on iOS."
              : "Add MyGram to your device for a faster app-like experience."}
          </p>
          <div className="mt-3 flex gap-2">
            {!iosSafari ? (
              <Button type="button" size="sm" onClick={handleInstall}>
                Install
              </Button>
            ) : null}
            <Button type="button" size="sm" variant="outline" onClick={dismiss}>
              Not now
            </Button>
          </div>
        </div>
        <Button type="button" size="icon" variant="ghost" onClick={dismiss} aria-label="Dismiss install prompt">
          <X className="h-4 w-4" aria-hidden="true" />
        </Button>
      </div>
    </div>
  );
}
