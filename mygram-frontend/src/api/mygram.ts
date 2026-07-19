import axios, { type AxiosResponse } from "axios";

import { http } from "@/api/http";
import type {
  AdminStats,
  AdminUserUpdatePayload,
  AdminUsersQuery,
  AdminUsersResponse,
  AuthUser,
  BanUserPayload,
  Comment,
  CommentPayload,
  CurrentUserResponse,
  LoginPayload,
  LoginResponse,
  Photo,
  PhotoPayload,
  ProfileUpdatePayload,
  PushSubscriptionPayload,
  PushVapidPublicKeyResponse,
  PublicOpenAPISpec,
  RegisterPayload,
  SocialMedia,
  SocialMediaPayload,
  UploadPhotoResponse,
} from "@/api/types";

async function emptyListOn404<T>(request: Promise<AxiosResponse<T[]>>) {
  try {
    const response = await request;
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 404) {
      return [];
    }

    throw error;
  }
}

export const mygramApi = {
  register: async (payload: RegisterPayload) => {
    const response = await http.post<AuthUser>("/api/v1/auth/register", payload);
    return response.data;
  },

  login: async (payload: LoginPayload) => {
    const response = await http.post<LoginResponse>("/api/v1/auth/login", payload);
    return response.data;
  },

  me: async () => {
    const response = await http.get<CurrentUserResponse>("/api/v1/me");
    return response.data.user;
  },

  updateMe: async (payload: ProfileUpdatePayload) => {
    const response = await http.patch<CurrentUserResponse>("/api/v1/me", payload);
    return response.data.user;
  },

  listPhotos: () => emptyListOn404<Photo>(http.get("/api/v1/photos")),

  getPhoto: async (photoId: number) => {
    const response = await http.get<Photo>(`/api/v1/photos/${photoId}`);
    return response.data;
  },

  createPhoto: async (payload: PhotoPayload) => {
    const response = await http.post<Photo>("/api/v1/photos", payload);
    return response.data;
  },

  updatePhoto: async (photoId: number, payload: PhotoPayload) => {
    const response = await http.put<Photo>(`/api/v1/photos/${photoId}`, payload);
    return response.data;
  },

  uploadPhotoImage: async (file: File) => {
    const formData = new FormData();
    formData.append("file", file);

    const response = await http.post<UploadPhotoResponse>(
      "/api/v1/uploads/photos",
      formData,
      {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      },
    );
    return response.data;
  },

  deletePhoto: async (photoId: number) => {
    await http.delete(`/api/v1/photos/${photoId}`);
  },

  getPushVapidPublicKey: async () => {
    const response = await http.get<PushVapidPublicKeyResponse>(
      "/api/v1/push/vapid-public-key",
    );
    return response.data;
  },

  savePushSubscription: async (payload: PushSubscriptionPayload) => {
    await http.post("/api/v1/push/subscriptions", payload);
  },

  deletePushSubscription: async (endpoint: string) => {
    await http.delete("/api/v1/push/subscriptions", {
      data: { endpoint },
    });
  },

  listComments: () => emptyListOn404<Comment>(http.get("/api/v1/comments")),

  listCommentsForPhoto: (photoId: number) =>
    emptyListOn404<Comment>(http.get(`/api/v1/photos/${photoId}/comments`)),

  createComment: async (photoId: number, payload: CommentPayload) => {
    const response = await http.post<Comment>(
      `/api/v1/photos/${photoId}/comments`,
      payload,
    );
    return response.data;
  },

  updateComment: async (commentId: number, payload: CommentPayload) => {
    const response = await http.put<Comment>(`/api/v1/comments/${commentId}`, payload);
    return response.data;
  },

  deleteComment: async (commentId: number) => {
    await http.delete(`/api/v1/comments/${commentId}`);
  },

  listSocialMedia: () => emptyListOn404<SocialMedia>(http.get("/api/v1/social-media")),

  createSocialMedia: async (payload: SocialMediaPayload) => {
    const response = await http.post<SocialMedia>("/api/v1/social-media", payload);
    return response.data;
  },

  updateSocialMedia: async (socialMediaId: number, payload: SocialMediaPayload) => {
    const response = await http.put<SocialMedia>(
      `/api/v1/social-media/${socialMediaId}`,
      payload,
    );
    return response.data;
  },

  deleteSocialMedia: async (socialMediaId: number) => {
    await http.delete(`/api/v1/social-media/${socialMediaId}`);
  },

  getAdminStats: async () => {
    const response = await http.get<AdminStats>("/api/v1/admin/stats");
    return response.data;
  },

  listAdminUsers: async (query: AdminUsersQuery) => {
    const response = await http.get<AdminUsersResponse>("/api/v1/admin/users", {
      params: query,
    });
    return response.data;
  },

  updateAdminUser: async (userId: number, payload: AdminUserUpdatePayload) => {
    const response = await http.patch<AuthUser>(`/api/v1/admin/users/${userId}`, payload);
    return response.data;
  },

  banAdminUser: async (userId: number, payload: BanUserPayload) => {
    const response = await http.post<AuthUser>(
      `/api/v1/admin/users/${userId}/ban`,
      payload,
    );
    return response.data;
  },

  unbanAdminUser: async (userId: number) => {
    const response = await http.post<AuthUser>(`/api/v1/admin/users/${userId}/unban`);
    return response.data;
  },

  deleteAdminUser: async (userId: number) => {
    await http.delete(`/api/v1/admin/users/${userId}`);
  },

  getPublicOpenAPI: async () => {
    const response = await http.get<PublicOpenAPISpec>("/openapi/public.json");
    return response.data;
  },
};
