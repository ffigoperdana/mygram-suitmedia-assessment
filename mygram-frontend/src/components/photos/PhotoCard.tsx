import { FormEvent, useState } from "react";
import { Link } from "react-router-dom";
import { formatDistanceToNow } from "date-fns";
import { MessageCircle, Pencil, Trash2, X } from "lucide-react";
import { toast } from "sonner";

import type { Photo } from "@/api/types";
import { getApiErrorMessage } from "@/api/http";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { ImageUploadField } from "@/components/photos/ImageUploadField";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { ownerId } from "@/lib/utils";
import { useDeletePhoto, useUpdatePhoto, useUploadPhotoImage } from "@/hooks/use-photos";
import { useAuthStore } from "@/stores/auth-store";

export function PhotoCard({ photo }: { photo: Photo }) {
  const [editing, setEditing] = useState(false);
  const [title, setTitle] = useState(photo.title);
  const [caption, setCaption] = useState(photo.caption ?? "");
  const [photoUrl, setPhotoUrl] = useState(photo.photo_url);
  const [imageFile, setImageFile] = useState<File | null>(null);
  const currentUser = useAuthStore((state) => state.user);
  const updatePhoto = useUpdatePhoto(photo.id);
  const uploadPhotoImage = useUploadPhotoImage();
  const deletePhoto = useDeletePhoto();
  const isOwner = currentUser?.id === ownerId(photo);
  const isSaving = updatePhoto.isPending || uploadPhotoImage.isPending;

  const createdAt = photo.created_at
    ? formatDistanceToNow(new Date(photo.created_at), { addSuffix: true })
    : null;

  async function handleUpdate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    try {
      let nextPhotoUrl = photoUrl.trim();
      if (imageFile) {
        const upload = await uploadPhotoImage.mutateAsync(imageFile);
        nextPhotoUrl = upload.url;
        setPhotoUrl(upload.url);
      }

      await updatePhoto.mutateAsync({ title, caption, photo_url: nextPhotoUrl });
      setImageFile(null);
      setEditing(false);
      toast.success("Photo updated");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  function cancelEditing() {
    setTitle(photo.title);
    setCaption(photo.caption ?? "");
    setPhotoUrl(photo.photo_url);
    setImageFile(null);
    setEditing(false);
  }

  async function handleDelete() {
    if (!window.confirm("Delete this photo?")) {
      return;
    }

    try {
      await deletePhoto.mutateAsync(photo.id);
      toast.success("Photo deleted");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  return (
    <Card className="overflow-hidden">
      <Link to={`/photos/${photo.id}`} className="block">
        <div className="aspect-[4/3] w-full overflow-hidden bg-muted">
          <img
            src={photo.photo_url}
            alt={photo.title}
            className="h-full w-full object-cover transition-transform duration-200 hover:scale-[1.02]"
            loading="lazy"
          />
        </div>
      </Link>
      <div className="space-y-3 p-4">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <Link to={`/photos/${photo.id}`} className="font-semibold hover:underline">
              {photo.title}
            </Link>
            {createdAt ? <p className="text-xs text-muted-foreground">{createdAt}</p> : null}
          </div>
          <Badge>Photo</Badge>
        </div>
        {photo.caption ? (
          <p className="line-clamp-3 text-sm text-muted-foreground">{photo.caption}</p>
        ) : null}

        {editing ? (
          <form onSubmit={handleUpdate} className="grid gap-2 rounded-md border bg-background p-3">
            <Input
              value={title}
              onChange={(event) => setTitle(event.target.value)}
              aria-label="Photo title"
              required
            />
            <ImageUploadField
              idPrefix={`photo-${photo.id}`}
              imageUrl={photoUrl}
              imageFile={imageFile}
              onImageUrlChange={setPhotoUrl}
              onImageFileChange={setImageFile}
              disabled={isSaving}
              required
            />
            <Textarea
              value={caption}
              onChange={(event) => setCaption(event.target.value)}
              aria-label="Photo caption"
            />
            <div className="flex flex-wrap gap-2">
              <Button type="submit" size="sm" disabled={isSaving}>
                {uploadPhotoImage.isPending ? "Uploading" : updatePhoto.isPending ? "Saving" : "Save"}
              </Button>
              <Button type="button" size="sm" variant="outline" onClick={cancelEditing}>
                <X className="mr-2 h-4 w-4" aria-hidden="true" />
                Cancel
              </Button>
            </div>
          </form>
        ) : null}

        <div className="flex flex-wrap items-center justify-between gap-2">
          <Link
            to={`/photos/${photo.id}`}
            className="inline-flex items-center gap-2 text-sm font-medium text-primary hover:underline"
          >
            <MessageCircle className="h-4 w-4" aria-hidden="true" />
            Comments
          </Link>
          {isOwner ? (
            <div className="flex gap-1">
              <Button
                type="button"
                size="icon"
                variant="ghost"
                onClick={() => setEditing((current) => !current)}
                aria-label="Edit photo"
              >
                <Pencil className="h-4 w-4" aria-hidden="true" />
              </Button>
              <Button
                type="button"
                size="icon"
                variant="ghost"
                onClick={handleDelete}
                aria-label="Delete photo"
              >
                <Trash2 className="h-4 w-4" aria-hidden="true" />
              </Button>
            </div>
          ) : null}
        </div>
      </div>
    </Card>
  );
}
