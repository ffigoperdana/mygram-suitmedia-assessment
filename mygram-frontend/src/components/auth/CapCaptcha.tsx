import Cap, { type CapErrorEvent, type CapProgressEvent } from "cap-widget";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Loader2, RefreshCcw, ShieldCheck, ShieldQuestion } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";

type CapCaptchaProps = {
  value: string;
  onChange: (token: string) => void;
};

type VerificationState = "idle" | "verifying" | "verified" | "error";

function getCapConfig() {
  return {
    baseUrl: import.meta.env.VITE_CAP_BASE_URL ?? "",
    siteKey: import.meta.env.VITE_CAP_SITE_KEY ?? "",
    enabled: import.meta.env.VITE_CAP_ENABLED === "true",
  };
}

export function CapCaptcha({ value, onChange }: CapCaptchaProps) {
  const [open, setOpen] = useState(false);
  const [verificationState, setVerificationState] = useState<VerificationState>("idle");
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState("");
  const activeCapRef = useRef<Cap | null>(null);
  const runIdRef = useRef(0);
  const capConfig = getCapConfig();
  const apiEndpoint = useMemo(() => {
    if (!capConfig.enabled || !capConfig.baseUrl || !capConfig.siteKey) {
      return "";
    }

    return `${capConfig.baseUrl.replace(/\/$/, "")}/${encodeURIComponent(capConfig.siteKey)}/`;
  }, [capConfig.baseUrl, capConfig.enabled, capConfig.siteKey]);

  const cleanupActiveCap = useCallback(() => {
    const activeCap = activeCapRef.current;
    activeCapRef.current = null;
    if (!activeCap) {
      return;
    }

    try {
      activeCap.reset();
      activeCap.widget.remove();
    } catch {
      // The Cap widget owns its internal cleanup. Ignore post-solve cleanup races.
    }
  }, []);

  useEffect(() => {
    return () => {
      runIdRef.current += 1;
      cleanupActiveCap();
    };
  }, [cleanupActiveCap]);

  const startVerification = useCallback(async () => {
    if (!apiEndpoint) {
      setError("Captcha is enabled but frontend Cap configuration is incomplete.");
      return;
    }

    const runId = runIdRef.current + 1;
    runIdRef.current = runId;
    cleanupActiveCap();
    onChange("");
    setOpen(true);
    setError("");
    setProgress(0);
    setVerificationState("verifying");

    try {
      const cap = new Cap({
        apiEndpoint,
        "data-cap-worker-count": "2",
        "data-cap-hidden-field-name": "captcha_token",
      });
      activeCapRef.current = cap;

      cap.addEventListener("progress", (event: CapProgressEvent) => {
        if (runIdRef.current === runId) {
          setProgress(Math.max(0, Math.min(100, Math.round(event.detail.progress))));
        }
      });
      cap.addEventListener("error", (event: CapErrorEvent) => {
        if (runIdRef.current === runId) {
          setError(event.detail.message || "Captcha verification could not start.");
          setVerificationState("error");
        }
      });

      const result = await cap.solve();
      if (runIdRef.current !== runId) {
        return;
      }

      if (!result.success || !result.token) {
        throw new Error("Captcha verification did not return a token.");
      }

      onChange(result.token);
      setProgress(100);
      setVerificationState("verified");
      window.setTimeout(() => {
        if (runIdRef.current === runId) {
          setOpen(false);
          cleanupActiveCap();
        }
      }, 450);
    } catch (solveError) {
      if (runIdRef.current !== runId) {
        return;
      }

      setProgress(0);
      setVerificationState("error");
      setError(
        solveError instanceof Error
          ? solveError.message
          : "Captcha verification could not start.",
      );
      onChange("");
      cleanupActiveCap();
    }
  }, [apiEndpoint, cleanupActiveCap, onChange]);

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen && verificationState !== "verifying") {
      cleanupActiveCap();
    }
  }

  if (!capConfig.enabled) {
    return null;
  }

  if (!apiEndpoint) {
    return (
      <div className="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
        Captcha is enabled but frontend Cap configuration is incomplete.
      </div>
    );
  }

  return (
    <div className="grid gap-2">
      <Label>Captcha</Label>
      <div className="flex items-center justify-between gap-3 rounded-md border bg-background px-3 py-2">
        <p
          className={
            value
              ? "inline-flex min-w-0 items-center gap-2 text-sm font-medium text-primary"
              : "inline-flex min-w-0 items-center gap-2 text-sm text-muted-foreground"
          }
        >
          {value ? (
            <>
              <ShieldCheck className="h-4 w-4 shrink-0" aria-hidden="true" />
              <span>Captcha verified</span>
            </>
          ) : (
            <>
              <ShieldQuestion className="h-4 w-4 shrink-0" aria-hidden="true" />
              <span>Verification required</span>
            </>
          )}
        </p>
        <Button
          type="button"
          size="sm"
          variant={value ? "outline" : "secondary"}
          onClick={() => {
            void startVerification();
          }}
        >
          {value ? "Verify again" : "Verify human"}
        </Button>
      </div>
      {error ? <p className="text-sm text-destructive">{error}</p> : null}

      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Verify you are human</DialogTitle>
            <DialogDescription>
              Complete the check once, then continue signing in to MyGram.
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4">
            <div className="rounded-md border bg-background p-4">
              <div className="flex items-center gap-3">
                <span className="inline-flex h-10 w-10 items-center justify-center rounded-md bg-primary/10 text-primary">
                  {verificationState === "verified" ? (
                    <ShieldCheck className="h-5 w-5" aria-hidden="true" />
                  ) : (
                    <Loader2 className="h-5 w-5 animate-spin" aria-hidden="true" />
                  )}
                </span>
                <div className="min-w-0">
                  <p className="text-sm font-medium">
                    {verificationState === "verified"
                      ? "Verification complete"
                      : verificationState === "error"
                        ? "Verification failed"
                        : "Checking your browser"}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    {verificationState === "verified"
                      ? "You can continue signing in."
                      : verificationState === "error"
                        ? "Refresh the check and try again."
                        : "This usually takes a few seconds."}
                  </p>
                </div>
              </div>
              <div
                className="mt-4 h-2 overflow-hidden rounded-full bg-muted"
                role="progressbar"
                aria-valuemin={0}
                aria-valuemax={100}
                aria-valuenow={progress}
              >
                <div
                  className="h-full rounded-full bg-primary transition-all"
                  style={{ width: `${progress}%` }}
                />
              </div>
            </div>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <Button
              type="button"
              variant="ghost"
              className="justify-self-start px-0"
              disabled={verificationState === "verifying"}
              onClick={() => {
                void startVerification();
              }}
            >
              <RefreshCcw className="mr-2 h-4 w-4" aria-hidden="true" />
              Refresh check
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
