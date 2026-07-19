import { FormEvent, useState } from "react";
import { toast } from "sonner";

import { getApiErrorMessage } from "@/api/http";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { ImageUploadField } from "@/components/photos/ImageUploadField";
import { useCreatePhoto, useUploadPhotoImage } from "@/hooks/use-photos";

export function PhotoComposer() {
  const [title, setTitle] = useState("");
  const [caption, setCaption] = useState("");
  const [photoUrl, setPhotoUrl] = useState("");
  const [imageFile, setImageFile] = useState<File | null>(null);
  const createPhoto = useCreatePhoto();
  const uploadPhotoImage = useUploadPhotoImage();
  const isSubmitting = createPhoto.isPending || uploadPhotoImage.isPending;

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    try {
      let nextPhotoUrl = photoUrl.trim();
      if (imageFile) {
        const upload = await uploadPhotoImage.mutateAsync(imageFile);
        nextPhotoUrl = upload.url;
      }

      if (!nextPhotoUrl) {
        toast.error("Add an image URL or upload an image.");
        return;
      }

      await createPhoto.mutateAsync({
        title,
        caption,
        photo_url: nextPhotoUrl,
      });
      setTitle("");
      setCaption("");
      setPhotoUrl("");
      setImageFile(null);
      toast.success("Photo posted");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>New Photo</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="grid gap-4">
          <div className="grid gap-2">
            <Label htmlFor="photo-title">Title</Label>
            <Input
              id="photo-title"
              value={title}
              onChange={(event) => setTitle(event.target.value)}
              required
            />
          </div>
          <div className="grid gap-2">
            <ImageUploadField
              idPrefix="photo"
              imageUrl={photoUrl}
              imageFile={imageFile}
              onImageUrlChange={setPhotoUrl}
              onImageFileChange={setImageFile}
              disabled={isSubmitting}
              required
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="photo-caption">Caption</Label>
            <Textarea
              id="photo-caption"
              value={caption}
              onChange={(event) => setCaption(event.target.value)}
            />
          </div>
          <Button type="submit" disabled={isSubmitting}>
            {uploadPhotoImage.isPending ? "Uploading" : createPhoto.isPending ? "Posting" : "Post"}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
