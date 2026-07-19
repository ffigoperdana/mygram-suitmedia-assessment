import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";
import {
  BookOpen,
  Camera,
  Home,
  LayoutDashboard,
  Link2,
  LogOut,
  UserRound,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { PWANotificationButton } from "@/components/pwa/PWANotificationButton";
import { cn } from "@/lib/utils";
import { useCurrentUser } from "@/hooks/use-auth";
import { useAuthStore } from "@/stores/auth-store";

const baseNavItems = [
  { to: "/feed", label: "Feed", icon: Home },
  { to: "/social-links", label: "Links", icon: Link2 },
  { to: "/profile", label: "Profile", icon: UserRound },
  { to: "/docs", label: "API Docs", icon: BookOpen },
];

export function AppShell() {
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const logout = useAuthStore((state) => state.logout);
  useCurrentUser();

  const navItems =
    user?.role === "admin"
      ? [
          ...baseNavItems,
          { to: "/admin", label: "Admin", icon: LayoutDashboard },
        ]
      : baseNavItems;

  function handleLogout() {
    logout();
    navigate("/login", { replace: true });
  }

  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-20 border-b bg-background/95 backdrop-blur">
        <div className="container flex h-16 items-center justify-between gap-4">
          <Link to="/feed" className="flex min-w-0 items-center gap-2 font-semibold">
            <span className="grid h-9 w-9 shrink-0 place-items-center rounded-md bg-primary text-primary-foreground">
              <Camera className="h-5 w-5" aria-hidden="true" />
            </span>
            <span className="truncate">MyGram</span>
          </Link>

          <nav className="hidden items-center gap-1 lg:flex">
            {navItems.map((item) => (
              <NavItem key={item.to} item={item} />
            ))}
          </nav>

          <div className="flex items-center gap-3">
            <div className="hidden text-right text-sm sm:block">
              <div className="flex items-center justify-end gap-2">
                <p className="font-medium">{user?.username ?? "MyGram user"}</p>
                {user?.role === "admin" ? <Badge>Admin</Badge> : null}
              </div>
              <p className="max-w-48 truncate text-muted-foreground">{user?.email}</p>
            </div>
            <PWANotificationButton />
            <Button variant="outline" size="icon" onClick={handleLogout} aria-label="Log out">
              <LogOut className="h-4 w-4" aria-hidden="true" />
            </Button>
          </div>
        </div>

        <nav className="container flex h-12 items-center gap-1 overflow-x-auto border-t lg:hidden">
          {navItems.map((item) => (
            <NavItem key={item.to} item={item} compact />
          ))}
        </nav>
      </header>

      <main className="container py-6">
        <Outlet />
      </main>
    </div>
  );
}

type NavItemProps = {
  item: (typeof baseNavItems)[number];
  compact?: boolean;
};

function NavItem({ item, compact = false }: NavItemProps) {
  return (
    <NavLink
      to={item.to}
      className={({ isActive }) =>
        cn(
          "inline-flex h-10 items-center justify-center gap-2 rounded-md px-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground",
          compact && "min-w-fit flex-1 px-2",
          isActive && "bg-muted text-foreground",
        )
      }
    >
      <item.icon className="h-4 w-4 shrink-0" aria-hidden="true" />
      <span>{item.label}</span>
    </NavLink>
  );
}
