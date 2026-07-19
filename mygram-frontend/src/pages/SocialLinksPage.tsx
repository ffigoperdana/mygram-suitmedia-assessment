import { FormEvent, useState } from "react";
import { ExternalLink, Pencil, Trash2, X } from "lucide-react";
import { toast } from "sonner";

import type { SocialMedia } from "@/api/types";
import { getApiErrorMessage } from "@/api/http";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useDocumentTitle } from "@/hooks/use-document-title";
import {
  useCreateSocialMediaLink,
  useDeleteSocialMediaLink,
  useSocialMediaLinks,
  useUpdateSocialMediaLink,
} from "@/hooks/use-social-media";
import {
  supportedSocialPlatforms,
  validateSocialProfileInput,
} from "@/lib/social-links";

export function SocialLinksPage() {
  useDocumentTitle("Social Links | MyGram");
  const links = useSocialMediaLinks();
  const createLink = useCreateSocialMediaLink();
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const urlInput = event.currentTarget.elements.namedItem("social-url");
    if (urlInput instanceof HTMLInputElement && !validateSocialProfileInput(urlInput)) {
      return;
    }

    try {
      await createLink.mutateAsync({
        name,
        social_media_url: url,
      });
      setName("");
      setUrl("");
      toast.success("Social link saved");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[360px_minmax(0,1fr)]">
      <Card>
        <CardHeader>
          <CardTitle>New Link</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="social-name">Name</Label>
              <Input
                id="social-name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                maxLength={80}
                required
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="social-url">URL</Label>
              <Input
                id="social-url"
                name="social-url"
                type="url"
                inputMode="url"
                pattern="https?://.+"
                title={`Use a direct profile or channel URL from ${supportedSocialPlatforms}.`}
                value={url}
                onChange={(event) => {
                  event.target.setCustomValidity("");
                  setUrl(event.target.value);
                }}
                maxLength={300}
                required
              />
              <p className="text-xs text-muted-foreground">
                Direct profile/channel only: {supportedSocialPlatforms}.
              </p>
            </div>
            <Button type="submit" disabled={createLink.isPending}>
              {createLink.isPending ? "Saving" : "Save link"}
            </Button>
          </form>
        </CardContent>
      </Card>

      <section className="space-y-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-normal">Social Links</h1>
          <p className="text-sm text-muted-foreground">Saved profiles and channels</p>
        </div>
        <div className="grid gap-3">
          {links.data?.map((link) => (
            <SocialLinkItem key={link.id} link={link} />
          ))}
          {!links.isLoading && links.data?.length === 0 ? (
            <div className="rounded-lg border bg-card p-6 text-sm text-muted-foreground">
              No social links yet.
            </div>
          ) : null}
        </div>
      </section>
    </div>
  );
}

function SocialLinkItem({ link }: { link: SocialMedia }) {
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(link.name);
  const [url, setUrl] = useState(link.social_media_url);
  const updateLink = useUpdateSocialMediaLink();
  const deleteLink = useDeleteSocialMediaLink();

  async function handleUpdate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const urlInput = event.currentTarget.elements.namedItem("social-url");
    if (urlInput instanceof HTMLInputElement && !validateSocialProfileInput(urlInput)) {
      return;
    }

    try {
      await updateLink.mutateAsync({
        socialMediaId: link.id,
        payload: { name, social_media_url: url },
      });
      setEditing(false);
      toast.success("Social link updated");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  async function handleDelete() {
    if (!window.confirm("Delete this social link?")) {
      return;
    }

    try {
      await deleteLink.mutateAsync(link.id);
      toast.success("Social link deleted");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  if (editing) {
    return (
      <form onSubmit={handleUpdate} className="grid gap-3 rounded-lg border bg-card p-4">
        <Input
          value={name}
          onChange={(event) => setName(event.target.value)}
          aria-label="Social link name"
          maxLength={80}
          required
        />
        <Input
          name="social-url"
          type="url"
          inputMode="url"
          pattern="https?://.+"
          title={`Use a direct profile or channel URL from ${supportedSocialPlatforms}.`}
          value={url}
          onChange={(event) => {
            event.target.setCustomValidity("");
            setUrl(event.target.value);
          }}
          aria-label="Social link URL"
          maxLength={300}
          required
        />
        <p className="text-xs text-muted-foreground">
          Direct profile/channel only: {supportedSocialPlatforms}.
        </p>
        <div className="flex flex-wrap gap-2">
          <Button type="submit" size="sm" disabled={updateLink.isPending}>
            Save
          </Button>
          <Button type="button" size="sm" variant="outline" onClick={() => setEditing(false)}>
            <X className="mr-2 h-4 w-4" aria-hidden="true" />
            Cancel
          </Button>
        </div>
      </form>
    );
  }

  return (
    <div className="flex items-center justify-between gap-4 rounded-lg border bg-card p-4">
      <a
        href={link.social_media_url}
        target="_blank"
        rel="noreferrer"
        className="min-w-0 hover:underline"
      >
        <p className="font-medium">{link.name}</p>
        <p className="truncate text-sm text-muted-foreground">{link.social_media_url}</p>
      </a>
      <div className="flex shrink-0 items-center gap-1">
        <a
          href={link.social_media_url}
          target="_blank"
          rel="noreferrer"
          className="grid h-10 w-10 place-items-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          aria-label="Open social link"
        >
          <ExternalLink className="h-4 w-4" aria-hidden="true" />
        </a>
        <Button
          type="button"
          size="icon"
          variant="ghost"
          onClick={() => setEditing(true)}
          aria-label="Edit social link"
        >
          <Pencil className="h-4 w-4" aria-hidden="true" />
        </Button>
        <Button
          type="button"
          size="icon"
          variant="ghost"
          onClick={handleDelete}
          aria-label="Delete social link"
        >
          <Trash2 className="h-4 w-4" aria-hidden="true" />
        </Button>
      </div>
    </div>
  );
}
