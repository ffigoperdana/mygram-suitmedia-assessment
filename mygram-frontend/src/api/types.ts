export type ApiErrorBody = {
  error?: string;
  err?: string;
  message?: string;
  error_status?: string;
  error_message?: string;
};

export type UserRole = "user" | "admin";
export type UserStatus = "active" | "banned";

export type AuthUser = {
  id: number;
  username: string;
  email: string;
  age?: number;
  role: UserRole;
  status: UserStatus;
  banned_at?: string | null;
  ban_reason?: string;
  last_login_at?: string | null;
  last_seen_at?: string | null;
  created_at?: string | null;
  updated_at?: string | null;
};

export type RegisterPayload = {
  username: string;
  email: string;
  password: string;
  age: number;
};

export type LoginPayload = {
  email: string;
  password: string;
};

export type LoginResponse = {
  token: string;
  user: AuthUser;
};

export type CurrentUserResponse = {
  user: AuthUser;
};

export type ProfileUpdatePayload = {
  username?: string;
  email?: string;
  age?: number;
};

export type Photo = {
  id: number;
  title: string;
  caption?: string;
  photo_url: string;
  UserID?: number;
  user_id?: number;
  created_at?: string;
  updated_at?: string;
};

export type PhotoPayload = {
  title: string;
  caption?: string;
  photo_url: string;
};

export type UploadPhotoResponse = {
  url: string;
  key: string;
  bucket: string;
  content_type: string;
  size: number;
};

export type PushVapidPublicKeyResponse = {
  enabled: boolean;
  public_key?: string;
};

export type PushSubscriptionPayload = {
  endpoint: string;
  keys: {
    p256dh: string;
    auth: string;
  };
  user_agent?: string;
};

export type Comment = {
  id: number;
  message: string;
  PhotoID?: number;
  photo_id?: number;
  UserID?: number;
  user_id?: number;
  created_at?: string;
  updated_at?: string;
};

export type CommentPayload = {
  message: string;
};

export type SocialMedia = {
  id: number;
  name: string;
  social_media_url: string;
  UserID?: number;
  user_id?: number;
  created_at?: string;
  updated_at?: string;
};

export type SocialMediaPayload = {
  name: string;
  social_media_url: string;
};

export type AdminStats = {
  total_users: number;
  active_users: number;
  banned_users: number;
  admin_users: number;
  users_seen_last_24h: number;
  total_photos: number;
  total_comments: number;
  total_social_media: number;
  recent_users: AuthUser[];
  generated_at: string;
};

export type AdminUsersQuery = {
  page?: number;
  limit?: number;
  search?: string;
  role?: UserRole | "";
  status?: UserStatus | "";
};

export type AdminUsersResponse = {
  users: AuthUser[];
  total: number;
  page: number;
  limit: number;
};

export type AdminUserUpdatePayload = {
  username?: string;
  email?: string;
  age?: number;
  role?: UserRole;
  status?: UserStatus;
  ban_reason?: string;
};

export type BanUserPayload = {
  reason?: string;
};

export type PublicOpenAPISpec = {
  swagger?: string;
  info?: {
    title?: string;
    description?: string;
    version?: string;
  };
  paths: Record<string, Record<string, unknown>>;
};
