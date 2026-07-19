import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  ArrowRight,
  Camera,
  LockKeyhole,
  Mail,
  ShieldCheck,
  User,
} from "lucide-react";
import { toast } from "sonner";

import { getApiErrorMessage } from "@/api/http";
import { CapCaptcha } from "@/components/auth/CapCaptcha";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useRegister } from "@/hooks/use-auth";
import { useDocumentTitle } from "@/hooks/use-document-title";

function shouldRequireRegisterCaptcha() {
  return import.meta.env.VITE_CAP_ENABLED === "true";
}

export function RegisterPage() {
  useDocumentTitle("Register | MyGram");
  const navigate = useNavigate();
  const register = useRegister();
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [age, setAge] = useState("18");
  const [captchaToken, setCaptchaToken] = useState("");
  const captchaRequired = shouldRequireRegisterCaptcha();
  const submitDisabled = register.isPending || (captchaRequired && !captchaToken);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    try {
      await register.mutateAsync({
        username,
        email,
        password,
        age: Number(age),
        captcha_token: captchaToken || undefined,
      });
      toast.success("Account created");
      navigate("/login");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  return (
    <main className="min-h-screen bg-[radial-gradient(circle_at_top_left,hsl(var(--accent))_0,transparent_32%),linear-gradient(135deg,hsl(var(--background))_0%,#eef6f5_100%)] px-4 py-8">
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
              Create an account built for posting and API access.
            </h1>
            <p className="mt-4 text-base leading-7 text-muted-foreground">
              Register once, then publish photos from the app or through the documented API.
            </p>
            <div className="mt-8 flex items-center gap-3 text-sm text-muted-foreground">
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-card text-primary shadow-sm">
                <ShieldCheck className="h-4 w-4" aria-hidden="true" />
              </span>
              <span>
                {captchaRequired
                  ? "New accounts are protected by captcha before they reach the API."
                  : "Passwords are hashed before account data is stored."}
              </span>
            </div>
          </div>
        </section>

        <Card className="w-full max-w-md justify-self-center border-border/80 shadow-md">
          <CardHeader className="space-y-3 pb-4">
            <div className="flex items-start justify-between gap-4">
              <div>
                <CardTitle className="text-2xl leading-tight">Create MyGram Account</CardTitle>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">
                  Set up your profile and start posting.
                </p>
              </div>
              <span className="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
                <User className="h-5 w-5" aria-hidden="true" />
              </span>
            </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="username">Username</Label>
              <div className="relative">
                <User
                  className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                  aria-hidden="true"
                />
                <Input
                  id="username"
                  autoComplete="username"
                  value={username}
                  onChange={(event) => setUsername(event.target.value)}
                  className="pl-9"
                  required
                />
              </div>
            </div>
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
              <Label htmlFor="age">Age</Label>
              <Input
                id="age"
                type="number"
                min="9"
                max="100"
                value={age}
                onChange={(event) => setAge(event.target.value)}
                required
              />
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
                  autoComplete="new-password"
                  minLength={6}
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
              {register.isPending ? "Creating" : "Create account"}
              {!register.isPending ? (
                <ArrowRight className="h-4 w-4" aria-hidden="true" />
              ) : null}
            </Button>
          </form>
          <p className="mt-5 border-t pt-5 text-sm text-muted-foreground">
            Already registered?{" "}
            <Link to="/login" className="font-medium text-primary hover:underline">
              Sign in
            </Link>
          </p>
        </CardContent>
      </Card>
      </div>
    </main>
  );
}
