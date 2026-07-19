import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Bell, BellOff, Loader2 } from "lucide-react";
import { toast } from "sonner";

import { mygramApi } from "@/api/mygram";
import type { PushSubscriptionPayload } from "@/api/types";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

const blockedPermissionMessage =
  "Notification sudah diblokir oleh browser. Klik ikon di kiri address bar → Site settings → Notifications → Allow.";

type NotificationDeliveryMode = "push" | null;

function base64UrlToUint8Array(base64String: string) {
  const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
  const rawData = window.atob(base64);
  const output = new Uint8Array(rawData.length);

  for (let index = 0; index < rawData.length; index += 1) {
    output[index] = rawData.charCodeAt(index);
  }

  return output;
}

function pushSubscriptionToPayload(subscription: PushSubscription): PushSubscriptionPayload {
  const json = subscription.toJSON() as PushSubscriptionJSON & {
    keys?: {
      p256dh?: string;
      auth?: string;
    };
  };

  return {
    endpoint: json.endpoint ?? subscription.endpoint,
    keys: {
      p256dh: json.keys?.p256dh ?? "",
      auth: json.keys?.auth ?? "",
    },
    user_agent: navigator.userAgent.slice(0, 512),
  };
}

export function PWANotificationButton() {
  const user = useAuthStore((state) => state.user);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const [enabled, setEnabled] = useState(false);
  const [permission, setPermission] = useState<NotificationPermission>("default");
  const [isRequesting, setIsRequesting] = useState(false);
  const [deliveryMode, setDeliveryMode] = useState<NotificationDeliveryMode>(null);
  const isSyncingPushRef = useRef(false);
  const isSupported = typeof window !== "undefined" && "Notification" in window;
  const isSecure = typeof window !== "undefined" && window.isSecureContext;
  const pushSupported =
    typeof window !== "undefined" && "serviceWorker" in navigator && "PushManager" in window;

  const keys = useMemo(() => {
    const userId = user?.id ?? "anonymous";
    return {
      preference: `mygram:notifications:${userId}:enabled`,
    };
  }, [user?.id]);

  useEffect(() => {
    if (!isSupported || !user?.id) {
      setEnabled(false);
      return;
    }

    const currentPermission = Notification.permission;
    const storedEnabled = window.localStorage.getItem(keys.preference) === "true";
    setPermission(currentPermission);
    setEnabled(storedEnabled && currentPermission === "granted");

    if (currentPermission === "denied" && storedEnabled) {
      window.localStorage.setItem(keys.preference, "false");
    }
  }, [isSupported, keys.preference, user?.id]);

  const getServiceWorkerRegistration = useCallback(async () => {
    const existingRegistration = await navigator.serviceWorker.getRegistration("/");
    if (existingRegistration) {
      return existingRegistration;
    }

    return navigator.serviceWorker.register("/sw.js");
  }, []);

  const registerWebPushSubscription = useCallback(async () => {
    if (!pushSupported || isSyncingPushRef.current) {
      return false;
    }

    isSyncingPushRef.current = true;
    try {
      const vapid = await mygramApi.getPushVapidPublicKey();
      if (!vapid.enabled || !vapid.public_key) {
        return false;
      }

      const registration = await getServiceWorkerRegistration();
      let subscription = await registration.pushManager.getSubscription();

      if (!subscription) {
        subscription = await registration.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: base64UrlToUint8Array(vapid.public_key),
        });
      }

      await mygramApi.savePushSubscription(pushSubscriptionToPayload(subscription));
      setDeliveryMode("push");
      return true;
    } catch {
      setDeliveryMode(null);
      return false;
    } finally {
      isSyncingPushRef.current = false;
    }
  }, [getServiceWorkerRegistration, pushSupported]);

  const unregisterWebPushSubscription = useCallback(async () => {
    if (!pushSupported) {
      setDeliveryMode(null);
      return;
    }

    try {
      const registration = await navigator.serviceWorker.ready;
      const subscription = await registration.pushManager.getSubscription();

      if (subscription) {
        await mygramApi.deletePushSubscription(subscription.endpoint).catch(() => undefined);
        await subscription.unsubscribe().catch(() => undefined);
      }
    } finally {
      setDeliveryMode(null);
    }
  }, [pushSupported]);

  useEffect(() => {
    if (!isAuthenticated || !user?.id || !enabled || permission !== "granted") {
      setDeliveryMode(null);
      return;
    }

    void registerWebPushSubscription();
  }, [enabled, isAuthenticated, permission, registerWebPushSubscription, user?.id]);

  async function toggleNotifications() {
    if (!isSupported || !user?.id) {
      toast.error("This browser does not support notifications.");
      return;
    }

    if (!isSecure) {
      toast.error("Push notifications require HTTPS or localhost.");
      return;
    }

    if (!pushSupported) {
      toast.error("This browser does not support Web Push notifications.");
      return;
    }

    if (enabled) {
      setIsRequesting(true);
      try {
        await unregisterWebPushSubscription();
        window.localStorage.setItem(keys.preference, "false");
        setEnabled(false);
        toast.success("Notifications paused");
      } finally {
        setIsRequesting(false);
      }
      return;
    }

    setIsRequesting(true);
    try {
      let nextPermission = Notification.permission;

      if (nextPermission === "denied") {
        setPermission(nextPermission);
        toast.error(blockedPermissionMessage);
        return;
      }

      if (nextPermission === "default") {
        nextPermission = await Notification.requestPermission();
      }

      setPermission(nextPermission);
      if (nextPermission !== "granted") {
        toast.error("Notification permission was not granted.");
        return;
      }

      const registeredForPush = await registerWebPushSubscription();
      if (registeredForPush) {
        window.localStorage.setItem(keys.preference, "true");
        setEnabled(true);
        toast.success("Push notifications enabled");
      } else {
        window.localStorage.setItem(keys.preference, "false");
        setEnabled(false);
        toast.error("Push notifications are not ready. Check VAPID backend configuration.");
      }
    } finally {
      setIsRequesting(false);
    }
  }

  if (!isAuthenticated || !user?.id || !isSupported) {
    return null;
  }

  return (
    <Button
      type="button"
      variant="outline"
      size="icon"
      aria-label={enabled ? "Pause MyGram notifications" : "Enable MyGram notifications"}
      aria-pressed={enabled}
      title={
        enabled
          ? deliveryMode === "push"
            ? "Pause background push notifications"
            : "Pause notifications"
          : "Enable post and comment notifications"
      }
      className={cn(enabled && "border-primary text-primary")}
      onClick={() => {
        void toggleNotifications();
      }}
      disabled={isRequesting}
    >
      {isRequesting ? (
        <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
      ) : enabled ? (
        <Bell className="h-4 w-4" aria-hidden="true" />
      ) : (
        <BellOff className="h-4 w-4" aria-hidden="true" />
      )}
    </Button>
  );
}
