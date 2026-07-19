import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { mygramApi } from "@/api/mygram";
import type { ProfileUpdatePayload } from "@/api/types";
import { useAuthStore } from "@/stores/auth-store";

export const authKeys = {
  me: ["auth", "me"] as const,
};

export function useLogin() {
  return useMutation({
    mutationFn: mygramApi.login,
  });
}

export function useRegister() {
  return useMutation({
    mutationFn: mygramApi.register,
  });
}

export function useCurrentUser() {
  const token = useAuthStore((state) => state.token);
  const setUser = useAuthStore((state) => state.setUser);

  return useQuery({
    queryKey: authKeys.me,
    queryFn: async () => {
      const user = await mygramApi.me();
      setUser(user);
      return user;
    },
    enabled: Boolean(token),
    staleTime: 60_000,
    retry: false,
  });
}

export function useUpdateCurrentUser() {
  const queryClient = useQueryClient();
  const setUser = useAuthStore((state) => state.setUser);

  return useMutation({
    mutationFn: (payload: ProfileUpdatePayload) => mygramApi.updateMe(payload),
    onSuccess: (user) => {
      setUser(user);
      queryClient.setQueryData(authKeys.me, user);
    },
  });
}
