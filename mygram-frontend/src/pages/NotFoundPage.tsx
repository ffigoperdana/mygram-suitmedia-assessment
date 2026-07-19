import { Link } from "react-router-dom";

import { Button } from "@/components/ui/button";
import { useDocumentTitle } from "@/hooks/use-document-title";

export function NotFoundPage() {
  useDocumentTitle("Not Found | MyGram");
  return (
    <main className="grid min-h-screen place-items-center px-4">
      <div className="max-w-md text-center">
        <p className="text-sm font-medium text-primary">404</p>
        <h1 className="mt-2 text-3xl font-semibold tracking-normal">Page not found</h1>
        <p className="mt-3 text-muted-foreground">This MyGram page does not exist.</p>
        <Button asChild className="mt-6">
          <Link to="/feed">Go to feed</Link>
        </Button>
      </div>
    </main>
  );
}
