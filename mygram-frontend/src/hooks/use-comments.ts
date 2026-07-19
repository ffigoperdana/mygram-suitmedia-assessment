import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { mygramApi } from "@/api/mygram";
import type { CommentPayload } from "@/api/types";

export const commentKeys = {
  all: ["comments"] as const,
  byPhoto: (photoId: number) => ["comments", "photo", photoId] as const,
};

export function useCommentsForPhoto(photoId: number) {
  return useQuery({
    queryKey: commentKeys.byPhoto(photoId),
    queryFn: () => mygramApi.listCommentsForPhoto(photoId),
    enabled: Number.isFinite(photoId) && photoId > 0,
  });
}

export function useCreateComment(photoId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CommentPayload) => mygramApi.createComment(photoId, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: commentKeys.byPhoto(photoId) });
    },
  });
}

export function useUpdateComment(photoId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ commentId, payload }: { commentId: number; payload: CommentPayload }) =>
      mygramApi.updateComment(commentId, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: commentKeys.byPhoto(photoId) });
    },
  });
}

export function useDeleteComment(photoId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: mygramApi.deleteComment,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: commentKeys.byPhoto(photoId) });
    },
  });
}
