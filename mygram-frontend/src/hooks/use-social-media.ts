import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { mygramApi } from "@/api/mygram";
import type { SocialMediaPayload } from "@/api/types";

export const socialMediaKeys = {
  all: ["social-media"] as const,
};

export function useSocialMediaLinks() {
  return useQuery({
    queryKey: socialMediaKeys.all,
    queryFn: mygramApi.listSocialMedia,
  });
}

export function useCreateSocialMediaLink() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: mygramApi.createSocialMedia,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: socialMediaKeys.all }),
  });
}

export function useUpdateSocialMediaLink() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      socialMediaId,
      payload,
    }: {
      socialMediaId: number;
      payload: SocialMediaPayload;
    }) => mygramApi.updateSocialMedia(socialMediaId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: socialMediaKeys.all }),
  });
}

export function useDeleteSocialMediaLink() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: mygramApi.deleteSocialMedia,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: socialMediaKeys.all }),
  });
}
