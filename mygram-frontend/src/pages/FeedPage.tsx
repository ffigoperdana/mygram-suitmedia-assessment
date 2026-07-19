import { useMemo, useState } from "react";
import { Search } from "lucide-react";

import { PhotoCard } from "@/components/photos/PhotoCard";
import { PhotoComposer } from "@/components/photos/PhotoComposer";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { useDocumentTitle } from "@/hooks/use-document-title";
import { usePhotos } from "@/hooks/use-photos";
import { ownerId } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

export function FeedPage() {
  useDocumentTitle("Feed | MyGram");
  const photos = usePhotos();
  const currentUser = useAuthStore((state) => state.user);
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<"all" | "mine">("all");
  const filteredPhotos = useMemo(() => {
    const normalizedSearch = search.trim().toLowerCase();

    return (photos.data ?? []).filter((photo) => {
      const matchesOwner = filter === "all" || ownerId(photo) === currentUser?.id;
      const searchable = `${photo.title} ${photo.caption ?? ""}`.toLowerCase();
      const matchesSearch = !normalizedSearch || searchable.includes(normalizedSearch);

      return matchesOwner && matchesSearch;
    });
  }, [currentUser?.id, filter, photos.data, search]);

  return (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
      <section className="space-y-4">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold tracking-normal">Feed</h1>
            <p className="text-sm text-muted-foreground">Latest MyGram photos</p>
          </div>
          <div className="flex rounded-md border bg-background p-1">
            <Button
              type="button"
              size="sm"
              variant={filter === "all" ? "secondary" : "ghost"}
              onClick={() => setFilter("all")}
            >
              All
            </Button>
            <Button
              type="button"
              size="sm"
              variant={filter === "mine" ? "secondary" : "ghost"}
              onClick={() => setFilter("mine")}
            >
              Mine
            </Button>
          </div>
        </div>

        <div className="grid gap-2">
          <Label htmlFor="photo-search">Search photos</Label>
          <div className="relative">
            <Search
              className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
              aria-hidden="true"
            />
            <Input
              id="photo-search"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Search by title or caption"
              className="pl-9"
            />
          </div>
        </div>

        {photos.isLoading ? (
          <div className="grid gap-4 sm:grid-cols-2">
            <Skeleton className="h-80" />
            <Skeleton className="h-80" />
          </div>
        ) : null}

        {!photos.isLoading && photos.data?.length === 0 ? (
          <div className="rounded-lg border bg-card p-6 text-sm text-muted-foreground">
            No photos yet.
          </div>
        ) : null}

        {!photos.isLoading && photos.data?.length !== 0 && filteredPhotos.length === 0 ? (
          <div className="rounded-lg border bg-card p-6 text-sm text-muted-foreground">
            No photos match this view.
          </div>
        ) : null}

        <div className="grid gap-4 sm:grid-cols-2">
          {filteredPhotos.map((photo) => <PhotoCard key={photo.id} photo={photo} />)}
        </div>
      </section>

      <aside className="lg:sticky lg:top-24 lg:self-start">
        <PhotoComposer />
      </aside>
    </div>
  );
}
