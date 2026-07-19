import { Link } from "react-router-dom";
import { ArrowLeft, ExternalLink, Home } from "lucide-react";

import { apiDisplayBaseUrl } from "@/api/http";
import { Button } from "@/components/ui/button";
import { useDocumentTitle } from "@/hooks/use-document-title";

export function DocsSwaggerPage() {
  useDocumentTitle("Swagger | MyGram");
  const swaggerUrl = `${apiDisplayBaseUrl}/swagger-ui/index.html`;

  return (
    <main className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container flex min-h-16 flex-wrap items-center justify-between gap-3 py-3">
          <Button asChild variant="outline">
            <Link to="/docs">
              <ArrowLeft className="mr-2 h-4 w-4" aria-hidden="true" />
              Docs
            </Link>
          </Button>
          <div className="flex flex-wrap items-center gap-2">
            <Button asChild variant="outline">
              <Link to="/">
                <Home className="mr-2 h-4 w-4" aria-hidden="true" />
                Open app
              </Link>
            </Button>
            <Button asChild>
              <a href={swaggerUrl} target="_blank" rel="noreferrer">
                <ExternalLink className="mr-2 h-4 w-4" aria-hidden="true" />
                Open raw Swagger
              </a>
            </Button>
          </div>
        </div>
      </header>
      <section className="h-[calc(100vh-4.5rem)] bg-muted/30 p-2 sm:p-4">
        <div className="h-full overflow-hidden rounded-md border bg-background shadow-sm">
          <iframe
            title="MyGram public Swagger UI"
            src={swaggerUrl}
            className="h-full w-full border-0"
            loading="eager"
          />
        </div>
      </section>
    </main>
  );
}
