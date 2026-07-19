import { FormEvent, useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Pencil, Trash2, X } from "lucide-react";
import { toast } from "sonner";

import { getApiErrorMessage } from "@/api/http";
import { CommentList } from "@/components/comments/CommentList";
import { ImageUploadField } from "@/components/photos/ImageUploadField";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import { useDocumentTitle } from "@/hooks/use-document-title";
import { useDeletePhoto, usePhoto, useUpdatePhoto, useUploadPhotoImage } from "@/hooks/use-photos";
import { ownerId } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

export function PhotoDetailPage() {
  useDocumentTitle("Photo | MyGram");
  const navigate = useNavigate();
  const params = useParams();
  const photoId = Number(params.photoId);
  const photo = usePhoto(photoId);
  const updatePhoto = useUpdatePhoto(photoId);
  const uploadPhotoImage = useUploadPhotoImage();
  const deletePhoto = useDeletePhoto();
  const currentUser = useAuthStore((state) => state.user);
  const [editing, setEditing] = useState(false);
  const [title, setTitle] = useState("");
  const [caption, setCaption] = useState("");
  const [photoUrl, setPhotoUrl] = useState("");
  const [imageFile, setImageFile] = useState<File | null>(null);
  const isSaving = updatePhoto.isPending || uploadPhotoImage.isPending;

  useEffect(() => {
    if (photo.data) {
      setTitle(photo.data.title);
      setCaption(photo.data.caption ?? "");
      setPhotoUrl(photo.data.photo_url);
      setImageFile(null);
    }
  }, [photo.data]);

  if (photo.isLoading) {
    return <Skeleton className="h-[520px]" />;
  }

  if (!photo.data) {
    return (
      <div className="space-y-4">
        <Button asChild variant="outline">
          <Link to="/feed">
            <ArrowLeft className="mr-2 h-4 w-4" aria-hidden="true" />
            Back
          </Link>
        </Button>
        <div className="rounded-lg border bg-card p-6 text-sm text-muted-foreground">
          Photo not found.
        </div>
      </div>
    );
  }

  const isOwner = currentUser?.id === ownerId(photo.data);

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

  async function handleDelete() {
    if (!window.confirm("Delete this photo?")) {
      return;
    }

    try {
      await deletePhoto.mutateAsync(photoId);
      toast.success("Photo deleted");
      navigate("/feed", { replace: true });
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  function cancelEditing() {
    if (photo.data) {
      setTitle(photo.data.title);
      setCaption(photo.data.caption ?? "");
      setPhotoUrl(photo.data.photo_url);
    }
    setImageFile(null);
    setEditing(false);
  }

  return (
    <div className="mx-auto grid max-w-5xl gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Button asChild variant="outline" className="w-fit">
          <Link to="/feed">
            <ArrowLeft className="mr-2 h-4 w-4" aria-hidden="true" />
            Back
          </Link>
        </Button>
        {isOwner ? (
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={editing ? cancelEditing : () => setEditing(true)}
            >
              {editing ? (
                <X className="mr-2 h-4 w-4" aria-hidden="true" />
              ) : (
                <Pencil className="mr-2 h-4 w-4" aria-hidden="true" />
              )}
              {editing ? "Cancel" : "Edit"}
            </Button>
            <Button type="button" variant="destructive" onClick={handleDelete}>
              <Trash2 className="mr-2 h-4 w-4" aria-hidden="true" />
              Delete
            </Button>
          </div>
        ) : null}
      </div>

      {editing ? (
        <form onSubmit={handleUpdate} className="grid gap-3 rounded-lg border bg-card p-4">
          <Input
            value={title}
            onChange={(event) => setTitle(event.target.value)}
            aria-label="Photo title"
            required
          />
          <ImageUploadField
            idPrefix={`photo-detail-${photoId}`}
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
          <Button type="submit" className="w-fit" disabled={isSaving}>
            {uploadPhotoImage.isPending ? "Uploading" : updatePhoto.isPending ? "Saving" : "Save photo"}
          </Button>
        </form>
      ) : null}

      <article className="overflow-hidden rounded-lg border bg-card">
        <div className="aspect-[16/10] w-full bg-muted">
          <img
            src={photo.data.photo_url}
            alt={photo.data.title}
            className="h-full w-full object-cover"
          />
        </div>
        <div className="space-y-2 p-5">
          <h1 className="text-2xl font-semibold tracking-normal">{photo.data.title}</h1>
          {photo.data.caption ? (
            <p className="text-muted-foreground">{photo.data.caption}</p>
          ) : null}
        </div>
      </article>

      <CommentList photoId={photoId} />
    </div>
  );
}
