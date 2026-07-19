import { FormEvent, useMemo, useState } from "react";
import {
  Ban,
  CheckCircle2,
  Search,
  Shield,
  Trash2,
  UserCog,
  Users,
} from "lucide-react";
import { toast } from "sonner";

import type { AdminUserUpdatePayload, AuthUser, UserRole, UserStatus } from "@/api/types";
import { getApiErrorMessage } from "@/api/http";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { useDocumentTitle } from "@/hooks/use-document-title";
import {
  useAdminStats,
  useAdminUsers,
  useBanAdminUser,
  useDeleteAdminUser,
  useUnbanAdminUser,
  useUpdateAdminUser,
} from "@/hooks/use-admin";
import { useAuthStore } from "@/stores/auth-store";

export function AdminDashboardPage() {
  useDocumentTitle("Admin Dashboard | MyGram");
  const currentUser = useAuthStore((state) => state.user);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [role, setRole] = useState<UserRole | "">("");
  const [status, setStatus] = useState<UserStatus | "">("");
  const query = useMemo(
    () => ({ page, limit: 10, search, role, status }),
    [page, role, search, status],
  );
  const stats = useAdminStats();
  const users = useAdminUsers(query);
  const totalPages = Math.max(1, Math.ceil((users.data?.total ?? 0) / (users.data?.limit ?? 10)));

  function handleSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPage(1);
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-normal">Admin Dashboard</h1>
          <p className="text-sm text-muted-foreground">
            Monitor MyGram health and manage user access.
          </p>
        </div>
        <Badge className="bg-primary text-primary-foreground">
          <Shield className="mr-1 h-3.5 w-3.5" aria-hidden="true" />
          Admin
        </Badge>
      </div>

      {stats.isLoading ? (
        <div className="grid gap-3 md:grid-cols-4">
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={index} className="h-28" />
          ))}
        </div>
      ) : stats.data ? (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <Metric title="Total users" value={stats.data.total_users} icon={Users} />
          <Metric title="Active users" value={stats.data.active_users} icon={CheckCircle2} />
          <Metric title="Banned users" value={stats.data.banned_users} icon={Ban} />
          <Metric title="Photos" value={stats.data.total_photos} icon={UserCog} />
        </div>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>User Management</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <form onSubmit={handleSearch} className="grid gap-3 md:grid-cols-[1fr_160px_160px_auto]">
            <div className="grid gap-2">
              <Label htmlFor="admin-search">Search</Label>
              <div className="relative">
                <Search
                  className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                  aria-hidden="true"
                />
                <Input
                  id="admin-search"
                  value={search}
                  onChange={(event) => setSearch(event.target.value)}
                  placeholder="Username or email"
                  className="pl-9"
                />
              </div>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="admin-role">Role</Label>
              <select
                id="admin-role"
                value={role}
                onChange={(event) => setRole(event.target.value as UserRole | "")}
                className="h-10 rounded-md border border-input bg-background px-3 text-sm"
              >
                <option value="">All roles</option>
                <option value="user">User</option>
                <option value="admin">Admin</option>
              </select>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="admin-status">Status</Label>
              <select
                id="admin-status"
                value={status}
                onChange={(event) => setStatus(event.target.value as UserStatus | "")}
                className="h-10 rounded-md border border-input bg-background px-3 text-sm"
              >
                <option value="">All status</option>
                <option value="active">Active</option>
                <option value="banned">Banned</option>
              </select>
            </div>
            <Button type="submit" className="self-end">
              Apply
            </Button>
          </form>

          <div className="overflow-x-auto rounded-md border">
            <table className="min-w-[920px] w-full text-left text-sm">
              <thead className="bg-muted text-muted-foreground">
                <tr>
                  <th className="px-3 py-3 font-medium">User</th>
                  <th className="px-3 py-3 font-medium">Age</th>
                  <th className="px-3 py-3 font-medium">Role</th>
                  <th className="px-3 py-3 font-medium">Status</th>
                  <th className="px-3 py-3 font-medium">Reason</th>
                  <th className="px-3 py-3 font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.data?.users.map((user) => (
                  <AdminUserRow
                    key={user.id}
                    user={user}
                    currentUserId={currentUser?.id ?? 0}
                  />
                ))}
                {!users.isLoading && users.data?.users.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-3 py-8 text-center text-muted-foreground">
                      No users match these filters.
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>

          {users.isLoading ? <Skeleton className="h-40" /> : null}

          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className="text-sm text-muted-foreground">
              Page {page} of {totalPages}. Total {users.data?.total ?? 0} users.
            </p>
            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                disabled={page <= 1}
                onClick={() => setPage((current) => Math.max(1, current - 1))}
              >
                Previous
              </Button>
              <Button
                type="button"
                variant="outline"
                disabled={page >= totalPages}
                onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
              >
                Next
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

type MetricProps = {
  title: string;
  value: number;
  icon: typeof Users;
};

function Metric({ title, value, icon: Icon }: MetricProps) {
  return (
    <Card>
      <CardContent className="flex items-center justify-between gap-4 p-5">
        <div>
          <p className="text-sm text-muted-foreground">{title}</p>
          <p className="mt-1 text-2xl font-semibold">{value}</p>
        </div>
        <span className="grid h-10 w-10 place-items-center rounded-md bg-muted text-primary">
          <Icon className="h-5 w-5" aria-hidden="true" />
        </span>
      </CardContent>
    </Card>
  );
}

function AdminUserRow({
  user,
  currentUserId,
}: {
  user: AuthUser;
  currentUserId: number;
}) {
  const updateUser = useUpdateAdminUser();
  const banUser = useBanAdminUser();
  const unbanUser = useUnbanAdminUser();
  const deleteUser = useDeleteAdminUser();
  const [draft, setDraft] = useState<AdminUserUpdatePayload>({
    username: user.username,
    email: user.email,
    age: user.age,
    role: user.role,
    status: user.status,
    ban_reason: user.ban_reason ?? "",
  });
  const isSelf = user.id === currentUserId;
  const busy =
    updateUser.isPending || banUser.isPending || unbanUser.isPending || deleteUser.isPending;

  async function handleSave() {
    try {
      await updateUser.mutateAsync({ userId: user.id, payload: draft });
      toast.success("User updated");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  async function handleBan() {
    if (isSelf) {
      toast.error("You cannot ban your own admin account");
      return;
    }

    try {
      await banUser.mutateAsync({
        userId: user.id,
        payload: { reason: draft.ban_reason || "Manual admin action" },
      });
      toast.success("User banned");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  async function handleUnban() {
    try {
      await unbanUser.mutateAsync(user.id);
      toast.success("User unbanned");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  async function handleDelete() {
    if (isSelf) {
      toast.error("You cannot delete your own admin account");
      return;
    }

    if (!window.confirm(`Delete ${user.email}?`)) {
      return;
    }

    try {
      await deleteUser.mutateAsync(user.id);
      toast.success("User deleted");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  return (
    <tr className="border-t align-top">
      <td className="px-3 py-3">
        <div className="grid gap-2">
          <Input
            value={draft.username ?? ""}
            onChange={(event) => setDraft((current) => ({ ...current, username: event.target.value }))}
            aria-label={`Username for ${user.email}`}
          />
          <Input
            type="email"
            value={draft.email ?? ""}
            onChange={(event) => setDraft((current) => ({ ...current, email: event.target.value }))}
            aria-label={`Email for ${user.username}`}
          />
          {isSelf ? <Badge className="w-fit">Current admin</Badge> : null}
        </div>
      </td>
      <td className="px-3 py-3">
        <Input
          type="number"
          min="9"
          max="100"
          value={draft.age ?? ""}
          onChange={(event) =>
            setDraft((current) => ({ ...current, age: Number(event.target.value) }))
          }
          aria-label={`Age for ${user.email}`}
        />
      </td>
      <td className="px-3 py-3">
        <select
          value={draft.role}
          onChange={(event) =>
            setDraft((current) => ({ ...current, role: event.target.value as UserRole }))
          }
          disabled={isSelf}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
          aria-label={`Role for ${user.email}`}
        >
          <option value="user">User</option>
          <option value="admin">Admin</option>
        </select>
      </td>
      <td className="px-3 py-3">
        <select
          value={draft.status}
          onChange={(event) =>
            setDraft((current) => ({ ...current, status: event.target.value as UserStatus }))
          }
          disabled={isSelf}
          className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
          aria-label={`Status for ${user.email}`}
        >
          <option value="active">Active</option>
          <option value="banned">Banned</option>
        </select>
      </td>
      <td className="px-3 py-3">
        <Input
          value={draft.ban_reason ?? ""}
          onChange={(event) =>
            setDraft((current) => ({ ...current, ban_reason: event.target.value }))
          }
          placeholder="Optional reason"
          aria-label={`Ban reason for ${user.email}`}
        />
      </td>
      <td className="px-3 py-3">
        <div className="flex flex-wrap gap-2">
          <Button type="button" size="sm" onClick={handleSave} disabled={busy}>
            Save
          </Button>
          {user.status === "banned" ? (
            <Button type="button" size="sm" variant="outline" onClick={handleUnban} disabled={busy}>
              Unban
            </Button>
          ) : (
            <Button type="button" size="sm" variant="outline" onClick={handleBan} disabled={busy || isSelf}>
              Ban
            </Button>
          )}
          <Button
            type="button"
            size="icon"
            variant="ghost"
            onClick={handleDelete}
            disabled={busy || isSelf}
            aria-label={`Delete ${user.email}`}
          >
            <Trash2 className="h-4 w-4" aria-hidden="true" />
          </Button>
        </div>
      </td>
    </tr>
  );
}
