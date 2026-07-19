import { useEffect, useState, type ChangeEvent } from "react";
import { UploadCloud, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const maxImageUploadBytes = 4 * 1024 * 1024;
const allowedImageTypes = new Set(["image/jpeg", "image/png", "image/gif", "image/webp"]);

type ImageUploadFieldProps = {
  idPrefix: string;
  imageUrl: string;
  imageFile: File | null;
  onImageUrlChange: (value: string) => void;
  onImageFileChange: (file: File | null) => void;
  disabled?: boolean;
  required?: boolean;
};

export function ImageUploadField({
  idPrefix,
  imageUrl,
  imageFile,
  onImageUrlChange,
  onImageFileChange,
  disabled = false,
  required = false,
}: ImageUploadFieldProps) {
  const [previewUrl, setPreviewUrl] = useState("");
  const [fileInputKey, setFileInputKey] = useState(0);
  const imagePreview = previewUrl || imageUrl;

  useEffect(() => {
    if (!imageFile) {
      setPreviewUrl("");
      return;
    }

    const objectUrl = URL.createObjectURL(imageFile);
    setPreviewUrl(objectUrl);

    return () => URL.revokeObjectURL(objectUrl);
  }, [imageFile]);

  function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
    const input = event.currentTarget;
    const file = event.target.files?.[0] ?? null;
    input.setCustomValidity("");

    if (!file) {
      onImageFileChange(null);
      return;
    }

    if (!allowedImageTypes.has(file.type)) {
      input.setCustomValidity("Upload a JPG, PNG, GIF, or WebP image.");
      input.reportValidity();
      input.value = "";
      onImageFileChange(null);
      return;
    }

    if (file.size > maxImageUploadBytes) {
      input.setCustomValidity("Image must be 4 MB or smaller.");
      input.reportValidity();
      input.value = "";
      onImageFileChange(null);
      return;
    }

    onImageFileChange(file);
  }

  function clearSelectedFile() {
    onImageFileChange(null);
    setFileInputKey((value) => value + 1);
  }

  return (
    <div className="grid gap-3">
      <div className="grid gap-2">
        <Label htmlFor={`${idPrefix}-url`}>Image URL</Label>
        <Input
          id={`${idPrefix}-url`}
          type="url"
          inputMode="url"
          pattern="https?://.+"
          title="Use an http:// or https:// image URL."
          value={imageUrl}
          onChange={(event) => onImageUrlChange(event.target.value)}
          disabled={disabled || Boolean(imageFile)}
          required={required && !imageFile}
        />
      </div>

      <div className="grid gap-2 rounded-md border bg-muted/30 p-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <Label htmlFor={`${idPrefix}-file`}>Image file</Label>
            <p className="mt-1 text-xs text-muted-foreground">JPG, PNG, GIF, or WebP up to 4MB.</p>
          </div>
          {imageFile ? (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={clearSelectedFile}
              disabled={disabled}
            >
              <X className="mr-2 h-4 w-4" aria-hidden="true" />
              Clear
            </Button>
          ) : null}
        </div>
        <Input
          key={fileInputKey}
          id={`${idPrefix}-file`}
          type="file"
          accept="image/jpeg,image/png,image/gif,image/webp"
          onChange={handleFileChange}
          disabled={disabled}
        />
        {imageFile ? (
          <p className="flex min-w-0 items-center gap-2 text-xs text-muted-foreground">
            <UploadCloud className="h-4 w-4 shrink-0" aria-hidden="true" />
            <span className="truncate">{imageFile.name}</span>
          </p>
        ) : null}
      </div>

      {imagePreview ? (
        <div className="aspect-[16/9] overflow-hidden rounded-md border bg-muted">
          <img src={imagePreview} alt="" className="h-full w-full object-cover" />
        </div>
      ) : null}
    </div>
  );
}
