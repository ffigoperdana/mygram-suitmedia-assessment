import { FormEvent, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import {
  AlertTriangle,
  ArrowRight,
  BookOpen,
  Camera,
  LockKeyhole,
  Mail,
  ShieldCheck,
  UserPlus,
} from "lucide-react";
import { toast } from "sonner";

import { getApiErrorMessage } from "@/api/http";
import { CapCaptcha } from "@/components/auth/CapCaptcha";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useLogin } from "@/hooks/use-auth";
import { useDocumentTitle } from "@/hooks/use-document-title";
import { useAuthStore } from "@/stores/auth-store";

function shouldRequireLoginCaptcha() {
  return (
    import.meta.env.VITE_CAP_ENABLED === "true" &&
    import.meta.env.VITE_CAP_REQUIRED_ON_LOGIN !== "false"
  );
}

export function LoginPage() {
  useDocumentTitle("Login | MyGram");
  const navigate = useNavigate();
  const location = useLocation();
  const login = useLogin();
  const setSession = useAuthStore((state) => state.setSession);
  const notice = useAuthStore((state) => state.notice);
  const setNotice = useAuthStore((state) => state.setNotice);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [captchaToken, setCaptchaToken] = useState("");
  const captchaRequired = shouldRequireLoginCaptcha();
  const submitDisabled = login.isPending || (captchaRequired && !captchaToken);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setNotice(null);

    try {
      const response = await login.mutateAsync({
        email,
        password,
        captcha_token: captchaToken || undefined,
      });
      setSession(response.token, response.user);
      const redirectTo = (location.state as { from?: { pathname?: string } } | null)?.from
        ?.pathname;
      navigate(redirectTo ?? "/feed", { replace: true });
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  return (
    <main className="min-h-screen bg-[radial-gradient(circle_at_top_left,hsl(var(--secondary))_0,transparent_34%),linear-gradient(135deg,hsl(var(--background))_0%,#eef6f5_100%)] px-4 py-8">
      <div className="mx-auto grid min-h-[calc(100vh-4rem)] w-full max-w-5xl items-center gap-8 lg:grid-cols-[0.9fr_1fr]">
        <section className="hidden lg:block">
          <div className="max-w-sm">
            <div className="mb-5 inline-flex h-12 w-12 items-center justify-center rounded-md bg-primary text-primary-foreground shadow-sm">
              <Camera className="h-6 w-6" aria-hidden="true" />
            </div>
            <p className="mb-3 text-sm font-semibold uppercase tracking-normal text-primary">
              MyGram
            </p>
            <h1 className="text-4xl font-semibold leading-tight tracking-normal">
              Share photos, keep conversations moving.
            </h1>
            <p className="mt-4 text-base leading-7 text-muted-foreground">
              Sign in to manage your feed, comments, social links, and API access from one
              workspace.
            </p>
            <div className="mt-8 grid gap-3 text-sm text-muted-foreground">
              <div className="flex items-center gap-3">
                <span className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-card text-primary shadow-sm">
                  <ShieldCheck className="h-4 w-4" aria-hidden="true" />
                </span>
                <span>Captcha-protected auth with role-aware access.</span>
              </div>
              <div className="flex items-center gap-3">
                <span className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-card text-primary shadow-sm">
                  <BookOpen className="h-4 w-4" aria-hidden="true" />
                </span>
                <span>Public API documentation stays available for integrations.</span>
              </div>
            </div>
          </div>
        </section>

        <Card className="w-full max-w-md justify-self-center border-border/80 shadow-md">
          <CardHeader className="space-y-3 pb-4">
            <div className="flex items-start justify-between gap-4">
              <div>
                <CardTitle className="text-2xl leading-tight">Login to MyGram</CardTitle>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">
                  Continue to your feed and creator tools.
                </p>
              </div>
              <span className="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
                <Camera className="h-5 w-5" aria-hidden="true" />
              </span>
            </div>
        </CardHeader>
        <CardContent>
          {notice ? (
            <div className="mb-4 flex gap-3 rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
              <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
              <p>{notice.message}</p>
            </div>
          ) : null}

          <form onSubmit={handleSubmit} className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="email">Email</Label>
              <div className="relative">
                <Mail
                  className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                  aria-hidden="true"
                />
                <Input
                  id="email"
                  type="email"
                  autoComplete="email"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  className="pl-9"
                  required
                />
              </div>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="password">Password</Label>
              <div className="relative">
                <LockKeyhole
                  className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                  aria-hidden="true"
                />
                <Input
                  id="password"
                  type="password"
                  autoComplete="current-password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  className="pl-9"
                  required
                />
              </div>
            </div>
            {captchaRequired ? (
              <CapCaptcha value={captchaToken} onChange={setCaptchaToken} />
            ) : null}
            <Button type="submit" className="h-11 gap-2" disabled={submitDisabled}>
              {login.isPending ? "Signing in" : "Sign in"}
              {!login.isPending ? <ArrowRight className="h-4 w-4" aria-hidden="true" /> : null}
            </Button>
          </form>
          <div className="mt-5 grid gap-3 border-t pt-5 text-sm">
            <Link
              to="/register"
              className="inline-flex items-center gap-2 font-medium text-primary hover:underline"
            >
              <UserPlus className="h-4 w-4" aria-hidden="true" />
              New here? Create an account
            </Link>
            <Link
              to="/docs"
              className="inline-flex items-center gap-2 font-medium text-primary hover:underline"
            >
              <BookOpen className="h-4 w-4" aria-hidden="true" />
              Read API docs
            </Link>
          </div>
        </CardContent>
      </Card>
      </div>
    </main>
  );
}
