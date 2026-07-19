import { FormEvent, useEffect, useState } from "react";
import { formatDistanceToNow } from "date-fns";
import { Mail, Save, ShieldCheck, UserRound } from "lucide-react";
import { toast } from "sonner";

import { getApiErrorMessage } from "@/api/http";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useCurrentUser, useUpdateCurrentUser } from "@/hooks/use-auth";
import { useDocumentTitle } from "@/hooks/use-document-title";
import { useAuthStore } from "@/stores/auth-store";

export function ProfilePage() {
  useDocumentTitle("Profile | MyGram");
  const storedUser = useAuthStore((state) => state.user);
  const currentUser = useCurrentUser();
  const updateUser = useUpdateCurrentUser();
  const user = currentUser.data ?? storedUser;
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [age, setAge] = useState("18");
  const seen = user?.last_seen_at
    ? formatDistanceToNow(new Date(user.last_seen_at), { addSuffix: true })
    : null;

  useEffect(() => {
    if (!user) {
      return;
    }

    setUsername(user.username);
    setEmail(user.email);
    setAge(String(user.age ?? 18));
  }, [user]);

  async function handleUpdate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    try {
      await updateUser.mutateAsync({
        username,
        email,
        age: Number(age),
      });
      toast.success("Profile updated");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  return (
    <div className="mx-auto grid max-w-2xl gap-4">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between gap-4">
            <CardTitle>Profile</CardTitle>
            <Badge>{user?.role === "admin" ? "Admin" : "Member"}</Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-3">
            <span className="grid h-12 w-12 place-items-center rounded-md bg-primary text-primary-foreground">
              <UserRound className="h-6 w-6" aria-hidden="true" />
            </span>
            <div>
              <p className="font-semibold">{user?.username ?? "MyGram user"}</p>
              <p className="text-sm text-muted-foreground">User ID {user?.id ?? "-"}</p>
            </div>
          </div>

          <div className="grid gap-3 rounded-md border bg-background p-4">
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Mail className="h-4 w-4" aria-hidden="true" />
              {user?.email ?? "No email"}
            </div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <ShieldCheck className="h-4 w-4" aria-hidden="true" />
              Status: {user?.status ?? "unknown"}
            </div>
            {seen ? <p className="text-sm text-muted-foreground">Last seen {seen}</p> : null}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Edit Profile</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleUpdate} className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="profile-username">Username</Label>
              <Input
                id="profile-username"
                value={username}
                onChange={(event) => setUsername(event.target.value)}
                autoComplete="username"
                required
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="profile-email">Email</Label>
              <Input
                id="profile-email"
                type="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                autoComplete="email"
                required
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="profile-age">Age</Label>
              <Input
                id="profile-age"
                type="number"
                min="9"
                max="100"
                value={age}
                onChange={(event) => setAge(event.target.value)}
                required
              />
            </div>
            <Button type="submit" disabled={updateUser.isPending}>
              <Save className="mr-2 h-4 w-4" aria-hidden="true" />
              {updateUser.isPending ? "Saving" : "Save profile"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
