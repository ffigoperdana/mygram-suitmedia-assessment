import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { mygramApi } from "@/api/mygram";
import type { PhotoPayload } from "@/api/types";

export const photoKeys = {
  all: ["photos"] as const,
  detail: (photoId: number) => ["photos", photoId] as const,
};

export function usePhotos() {
  return useQuery({
    queryKey: photoKeys.all,
    queryFn: mygramApi.listPhotos,
  });
}

export function usePhoto(photoId: number) {
  return useQuery({
    queryKey: photoKeys.detail(photoId),
    queryFn: () => mygramApi.getPhoto(photoId),
    enabled: Number.isFinite(photoId) && photoId > 0,
  });
}

export function useCreatePhoto() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: mygramApi.createPhoto,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: photoKeys.all }),
  });
}

export function useUpdatePhoto(photoId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: PhotoPayload) => mygramApi.updatePhoto(photoId, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: photoKeys.all });
      queryClient.invalidateQueries({ queryKey: photoKeys.detail(photoId) });
    },
  });
}

export function useUploadPhotoImage() {
  return useMutation({
    mutationFn: mygramApi.uploadPhotoImage,
  });
}

export function useDeletePhoto() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: mygramApi.deletePhoto,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: photoKeys.all }),
  });
}
