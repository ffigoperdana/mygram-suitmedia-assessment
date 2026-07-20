import { FormEvent, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { ArrowRight, BookOpen, Copy, ExternalLink, Play } from "lucide-react";
import { toast } from "sonner";

import { apiDisplayBaseUrl } from "@/api/http";
import { mygramApi } from "@/api/mygram";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import { useDocumentTitle } from "@/hooks/use-document-title";

const registerCurlExample = `curl -X POST "${apiDisplayBaseUrl}/api/v1/auth/register" \\
  -H "Content-Type: application/json" \\
  -d '{"username":"mygram_user","email":"user@example.com","age":25,"password":"strong-password"}'`;

const loginCurlExample = `curl -X POST "${apiDisplayBaseUrl}/api/v1/auth/login" \\
  -H "Content-Type: application/json" \\
  -d '{"email":"user@example.com","password":"strong-password"}'`;

const loginResponseExample = `{
  "token": "<jwt>",
  "user": {
    "id": 1,
    "username": "mygram_user",
    "email": "user@example.com",
    "role": "user",
    "status": "active"
  }
}`;

const curlExample = `curl -X POST "${apiDisplayBaseUrl}/api/v1/photos" \\
  -H "Authorization: Bearer <your-jwt>" \\
  -H "Content-Type: application/json" \\
  -d '{"title":"Morning light","caption":"From the API","photo_url":"<url-from-upload-response>"}'`;

const uploadCurlExample = `curl -X POST "${apiDisplayBaseUrl}/api/v1/uploads/photos" \\
  -H "Authorization: Bearer <your-jwt>" \\
  -F "file=@photo.jpg"`;

const fetchExample = `await fetch("${apiDisplayBaseUrl}/api/v1/photos", {
  method: "POST",
  headers: {
    Authorization: "Bearer <your-jwt>",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    title: "Morning light",
    caption: "From the API",
    photo_url: "<url-from-upload-response>"
  })
});`;

export function ApiDocsPage() {
  useDocumentTitle("API Docs | MyGram");
  const spec = useQuery({
    queryKey: ["docs", "public-openapi"],
    queryFn: mygramApi.getPublicOpenAPI,
    retry: false,
  });
  const endpoints = Object.keys(spec.data?.paths ?? {});

  async function copy(value: string) {
    await navigator.clipboard.writeText(value);
    toast.success("Copied");
  }

  return (
    <main className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container flex h-16 items-center justify-between gap-4">
          <Link to="/docs" className="flex items-center gap-2 font-semibold">
            <span className="grid h-9 w-9 place-items-center rounded-md bg-primary text-primary-foreground">
              <BookOpen className="h-5 w-5" aria-hidden="true" />
            </span>
            <span>MyGram API Docs</span>
          </Link>
          <div className="flex gap-2">
            <Button asChild variant="outline">
              <Link to="/docs/swagger">Swagger</Link>
            </Button>
            <Button asChild>
              <Link to="/login">Open app</Link>
            </Button>
          </div>
        </div>
      </header>

      <div className="container grid gap-8 py-8">
        <section className="grid min-w-0 gap-4 md:grid-cols-[minmax(0,1fr)_320px]">
          <div className="min-w-0">
            <h1 className="text-3xl font-semibold tracking-normal">Build with MyGram</h1>
            <p className="mt-3 max-w-2xl text-muted-foreground">
              Use the public API to register, sign in, post photos, comment on photos, and
              manage social links. Admin endpoints are intentionally excluded from these public
              docs.
            </p>
          </div>
          <Card>
            <CardContent className="p-5">
              <p className="text-sm text-muted-foreground">API base URL</p>
              <p className="mt-1 break-all font-medium">{apiDisplayBaseUrl}</p>
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="mt-4"
                onClick={() => copy(apiDisplayBaseUrl)}
              >
                <Copy className="mr-2 h-4 w-4" aria-hidden="true" />
                Copy URL
              </Button>
            </CardContent>
          </Card>
        </section>

        <section className="grid min-w-0 gap-4 md:grid-cols-3">
          <Step
            number="1"
            title="Register or login"
            body="Register or login with your account credentials to receive a JWT in the response."
          />
          <Step
            number="2"
            title="Send bearer token"
            body="Use Authorization: Bearer <jwt> on protected routes. Tokens expire after 24 hours."
          />
          <Step
            number="3"
            title="Post content"
            body="Upload an image first, use the returned URL as photo_url, then create photos, comments, and social links through /api/v1."
          />
        </section>

        <section className="grid min-w-0 gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
          <Card>
            <CardHeader>
              <CardTitle>Get a JWT</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm text-muted-foreground">
              <p>
                Register or log in through the MyGram app or API. After login, copy the
                <code className="mx-1 rounded bg-muted px-1 py-0.5 font-mono text-xs">
                  token
                </code>
                value and send it as
                <code className="mx-1 rounded bg-muted px-1 py-0.5 font-mono text-xs">
                  Authorization: Bearer &lt;jwt&gt;
                </code>
                on protected endpoints.
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Post via API</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm text-muted-foreground">
              <p>Recommended upload flow:</p>
              <ol className="list-decimal space-y-2 pl-5">
                <li>Login and copy the JWT from the response.</li>
                <li>
                  Upload a JPG, PNG, GIF, or WebP file up to 4 MB to
                  <code className="mx-1 rounded bg-muted px-1 py-0.5 font-mono text-xs">
                    /api/v1/uploads/photos
                  </code>
                  .
                </li>
                <li>
                  Use the upload response
                  <code className="mx-1 rounded bg-muted px-1 py-0.5 font-mono text-xs">
                    url
                  </code>
                  as
                  <code className="mx-1 rounded bg-muted px-1 py-0.5 font-mono text-xs">
                    photo_url
                  </code>
                  when posting to
                  <code className="mx-1 rounded bg-muted px-1 py-0.5 font-mono text-xs">
                    /api/v1/photos
                  </code>
                  .
                </li>
              </ol>
            </CardContent>
          </Card>
        </section>

        <section className="grid min-w-0 gap-4">
          <Card>
            <CardHeader>
              <CardTitle>Push notifications</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm text-muted-foreground">
              <p>
                MyGram uses the Web Push API, VAPID, and a Service Worker. It does not use
                SMTP or email notifications.
              </p>
              <p>
                The frontend fetches the VAPID public key from
                <code className="mx-1 rounded bg-muted px-1 py-0.5 font-mono text-xs">
                  /api/v1/push/vapid-public-key
                </code>
                , then stores the browser subscription at
                <code className="mx-1 rounded bg-muted px-1 py-0.5 font-mono text-xs">
                  /api/v1/push/subscriptions
                </code>
                . The backend only sends push messages to saved subscriptions.
              </p>
              <p>
                Android Chrome and desktop browsers can receive notifications after the
                user clicks the notification button and grants permission. On iOS, Web Push
                works only after MyGram is installed as a PWA and notifications are allowed.
              </p>
            </CardContent>
          </Card>
        </section>

        <section className="grid min-w-0 gap-4 lg:grid-cols-2">
          <Example title="cURL register" value={registerCurlExample} onCopy={copy} />
          <Example title="cURL login" value={loginCurlExample} onCopy={copy} />
          <Example title="Login response" value={loginResponseExample} onCopy={copy} />
        </section>

        <section className="grid min-w-0 gap-4 lg:grid-cols-2">
          <Example title="cURL photo post" value={curlExample} onCopy={copy} />
          <Example title="cURL image upload" value={uploadCurlExample} onCopy={copy} />
          <Example title="JavaScript fetch" value={fetchExample} onCopy={copy} />
        </section>

        <TryRequestConsole />

        <section className="grid min-w-0 gap-4 lg:grid-cols-[360px_minmax(0,1fr)]">
          <Card>
            <CardHeader>
              <CardTitle>Core endpoints</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <Endpoint method="POST" path="/api/v1/auth/register" />
              <Endpoint method="POST" path="/api/v1/auth/login" />
              <Endpoint method="GET/PATCH" path="/api/v1/me" />
              <Endpoint method="GET/POST" path="/api/v1/photos" />
              <Endpoint method="POST" path="/api/v1/uploads/photos" />
              <Endpoint method="GET/POST" path="/api/v1/photos/:photoId/comments" />
              <Endpoint method="GET/POST" path="/api/v1/social-media" />
              <Endpoint method="GET" path="/api/v1/push/vapid-public-key" />
              <Endpoint method="POST/DELETE" path="/api/v1/push/subscriptions" />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Public OpenAPI</CardTitle>
            </CardHeader>
            <CardContent>
              {spec.isLoading ? <Skeleton className="h-28" /> : null}
              {spec.data ? (
                <div className="space-y-4">
                  <p className="text-sm text-muted-foreground">
                    Loaded {endpoints.length} public paths from the backend filtered spec.
                  </p>
                  <div className="max-h-56 overflow-auto rounded-md border bg-background p-3">
                    {endpoints.map((endpoint) => (
                      <p key={endpoint} className="whitespace-nowrap font-mono text-xs leading-6">
                        {endpoint}
                      </p>
                    ))}
                  </div>
                </div>
              ) : null}
              {spec.isError ? (
                <p className="text-sm text-destructive">
                  Could not load public OpenAPI. Check that the backend is running.
                </p>
              ) : null}
              <div className="mt-4 flex flex-wrap gap-2">
                <Button asChild variant="outline">
                  <a href={`${apiDisplayBaseUrl}/openapi/public.json`} target="_blank" rel="noreferrer">
                    <ExternalLink className="mr-2 h-4 w-4" aria-hidden="true" />
                    Raw JSON
                  </a>
                </Button>
                <Button asChild>
                  <Link to="/docs/swagger">
                    Swagger UI
                    <ArrowRight className="ml-2 h-4 w-4" aria-hidden="true" />
                  </Link>
                </Button>
              </div>
            </CardContent>
          </Card>
        </section>
      </div>
    </main>
  );
}

type TryRoute = {
  label: string;
  method: "GET" | "POST" | "PATCH";
  path: string;
  body: string;
};

const tryRoutes: TryRoute[] = [
  {
    label: "Current user",
    method: "GET",
    path: "/api/v1/me",
    body: "",
  },
  {
    label: "List photos",
    method: "GET",
    path: "/api/v1/photos",
    body: "",
  },
  {
    label: "Create photo",
    method: "POST",
    path: "/api/v1/photos",
    body: JSON.stringify(
      {
        title: "API test photo",
        caption: "Created from the docs console",
        photo_url: "https://images.unsplash.com/photo-1500530855697-b586d89ba3ee",
      },
      null,
      2,
    ),
  },
  {
    label: "Update profile",
    method: "PATCH",
    path: "/api/v1/me",
    body: JSON.stringify(
      {
        username: "mygram_user",
        email: "user@example.com",
        age: 25,
      },
      null,
      2,
    ),
  },
  {
    label: "Create social link",
    method: "POST",
    path: "/api/v1/social-media",
    body: JSON.stringify(
      {
        name: "GitHub",
        social_media_url: "https://github.com/mygram_user",
      },
      null,
      2,
    ),
  },
];

function TryRequestConsole() {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [token, setToken] = useState("");
  const [body, setBody] = useState(tryRoutes[0].body);
  const [response, setResponse] = useState("");
  const [isSending, setIsSending] = useState(false);
  const selected = tryRoutes[selectedIndex];
  const targetUrl = `${apiDisplayBaseUrl}${selected.path}`;

  function selectRoute(index: number) {
    const nextRoute = tryRoutes[index];
    setSelectedIndex(index);
    setBody(nextRoute.body);
    setResponse("");
  }

  async function sendRequest(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSending(true);
    setResponse("");

    try {
      const result = await fetch(targetUrl, {
        method: selected.method,
        headers: {
          Authorization: `Bearer ${token}`,
          ...(selected.method !== "GET" ? { "Content-Type": "application/json" } : {}),
        },
        body: selected.method !== "GET" ? body : undefined,
      });
      const text = await result.text();
      let formatted = text;
      try {
        formatted = text ? JSON.stringify(JSON.parse(text), null, 2) : "";
      } catch {
        formatted = text;
      }
      setResponse(`${result.status} ${result.statusText}\n${formatted}`);
    } catch (error) {
      setResponse(error instanceof Error ? error.message : "Request failed");
    } finally {
      setIsSending(false);
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <CardTitle>Try request</CardTitle>
            <p className="mt-1 text-sm text-muted-foreground">
              Uses your JWT in this tab only and targets the configured API base URL. Paste the
              token value returned from /api/v1/auth/login.
            </p>
          </div>
          <Badge>{selected.method}</Badge>
        </div>
      </CardHeader>
      <CardContent>
        <form onSubmit={sendRequest} className="grid min-w-0 gap-4">
          <div className="grid gap-2">
            <Label htmlFor="try-route">Endpoint</Label>
            <select
              id="try-route"
              value={selectedIndex}
              onChange={(event) => selectRoute(Number(event.target.value))}
              className="h-10 rounded-md border border-input bg-background px-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              {tryRoutes.map((route, index) => (
                <option key={route.label} value={index}>
                  {route.method} {route.path}
                </option>
              ))}
            </select>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="try-token">JWT</Label>
            <Input
              id="try-token"
              type="password"
              autoComplete="off"
              value={token}
              onChange={(event) => setToken(event.target.value)}
              placeholder="Paste your JWT"
              required
            />
          </div>

          {selected.method !== "GET" ? (
            <div className="grid gap-2">
              <Label htmlFor="try-body">JSON body</Label>
              <Textarea
                id="try-body"
                value={body}
                onChange={(event) => setBody(event.target.value)}
                className="min-h-44 font-mono text-xs"
                required
              />
            </div>
          ) : null}

          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className="break-all font-mono text-xs text-muted-foreground">{targetUrl}</p>
            <Button type="submit" disabled={isSending}>
              <Play className="mr-2 h-4 w-4" aria-hidden="true" />
              {isSending ? "Sending" : "Send"}
            </Button>
          </div>
        </form>

        {response ? (
          <pre className="mt-4 max-h-80 overflow-auto rounded-md bg-foreground p-4 text-xs text-background">
            <code>{response}</code>
          </pre>
        ) : null}
      </CardContent>
    </Card>
  );
}

function Step({ number, title, body }: { number: string; title: string; body: string }) {
  return (
    <Card>
      <CardContent className="p-5">
        <span className="grid h-8 w-8 place-items-center rounded-md bg-primary text-sm font-semibold text-primary-foreground">
          {number}
        </span>
        <h2 className="mt-4 font-semibold">{title}</h2>
        <p className="mt-2 text-sm text-muted-foreground">{body}</p>
      </CardContent>
    </Card>
  );
}

function Example({
  title,
  value,
  onCopy,
}: {
  title: string;
  value: string;
  onCopy: (value: string) => void;
}) {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between gap-3">
          <CardTitle>{title}</CardTitle>
          <Button type="button" size="icon" variant="ghost" onClick={() => onCopy(value)} aria-label={`Copy ${title}`}>
            <Copy className="h-4 w-4" aria-hidden="true" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <pre className="overflow-auto rounded-md bg-foreground p-4 text-xs text-background">
          <code>{value}</code>
        </pre>
      </CardContent>
    </Card>
  );
}

function Endpoint({ method, path }: { method: string; path: string }) {
  return (
    <div className="min-w-0 overflow-x-auto rounded-md border bg-background">
      <div className="flex min-w-max items-center gap-3 px-3 py-2">
        <span className="w-28 shrink-0 rounded bg-muted px-2 py-1 text-center font-mono text-xs">
          {method}
        </span>
        <span className="whitespace-nowrap font-mono text-xs">{path}</span>
      </div>
    </div>
  );
}
