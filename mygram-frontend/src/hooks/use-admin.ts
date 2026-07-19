import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { mygramApi } from "@/api/mygram";
import type {
  AdminUserUpdatePayload,
  AdminUsersQuery,
  BanUserPayload,
} from "@/api/types";

export const adminKeys = {
  stats: ["admin", "stats"] as const,
  users: (query: AdminUsersQuery) => ["admin", "users", query] as const,
};

export function useAdminStats() {
  return useQuery({
    queryKey: adminKeys.stats,
    queryFn: mygramApi.getAdminStats,
  });
}

export function useAdminUsers(query: AdminUsersQuery) {
  return useQuery({
    queryKey: adminKeys.users(query),
    queryFn: () => mygramApi.listAdminUsers(query),
  });
}

export function useUpdateAdminUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      userId,
      payload,
    }: {
      userId: number;
      payload: AdminUserUpdatePayload;
    }) => mygramApi.updateAdminUser(userId, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin"] });
    },
  });
}

export function useBanAdminUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ userId, payload }: { userId: number; payload: BanUserPayload }) =>
      mygramApi.banAdminUser(userId, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin"] });
    },
  });
}

export function useUnbanAdminUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: mygramApi.unbanAdminUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin"] });
    },
  });
}

export function useDeleteAdminUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: mygramApi.deleteAdminUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin"] });
    },
  });
}
