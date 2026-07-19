const cacheName = "mygram-shell-v1";
const shellAssets = [
  "/",
  "/manifest.webmanifest",
  "/icons/mygram-icon.svg",
  "/icons/mygram-icon-192.png",
  "/icons/mygram-icon-512.png",
  "/icons/apple-touch-icon.png",
];

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(cacheName).then((cache) => cache.addAll(shellAssets)),
  );
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) =>
        Promise.all(keys.filter((key) => key !== cacheName).map((key) => caches.delete(key))),
      ),
  );
  self.clients.claim();
});

self.addEventListener("fetch", (event) => {
  const request = event.request;
  const url = new URL(request.url);

  if (url.origin !== self.location.origin || url.pathname.startsWith("/api/")) {
    return;
  }

  if (request.mode === "navigate") {
    event.respondWith(
      fetch(request).catch(() => caches.match("/")),
    );
    return;
  }

  event.respondWith(
    caches.match(request).then((cached) => cached ?? fetch(request)),
  );
});

self.addEventListener("message", (event) => {
  if (event.data?.type !== "MYGRAM_NOTIFICATION") {
    return;
  }

  const payload = event.data.payload ?? {};
  event.waitUntil(
    self.registration.showNotification(payload.title ?? "MyGram", {
      body: payload.body ?? "",
      icon: "/icons/mygram-icon-192.png",
      badge: "/icons/mygram-icon-192.png",
      tag: payload.tag,
      data: {
        url: payload.url ?? "/feed",
      },
    }),
  );
});

self.addEventListener("push", (event) => {
  let payload = {};

  if (event.data) {
    try {
      payload = event.data.json();
    } catch {
      payload = {
        body: event.data.text(),
      };
    }
  }

  event.waitUntil(
    self.registration.showNotification(payload.title ?? "MyGram", {
      body: payload.body ?? "",
      icon: "/icons/mygram-icon-192.png",
      badge: "/icons/mygram-icon-192.png",
      tag: payload.tag ?? "mygram-notification",
      data: {
        url: payload.url ?? "/feed",
      },
    }),
  );
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();
  const url = event.notification.data?.url ?? "/feed";

  event.waitUntil(
    self.clients.matchAll({ type: "window", includeUncontrolled: true }).then((clients) => {
      const targetUrl = new URL(url, self.location.origin).href;
      const matchingClient = clients.find((client) => client.url === targetUrl);

      if (matchingClient) {
        return matchingClient.focus();
      }

      return self.clients.openWindow(targetUrl);
    }),
  );
});
