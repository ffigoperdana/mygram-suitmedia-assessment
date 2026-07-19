import { expect, type Page, type Route, test } from "@playwright/test";

const corsHeaders = {
  "access-control-allow-origin": "*",
  "access-control-allow-methods": "GET,POST,PUT,PATCH,DELETE,OPTIONS",
  "access-control-allow-headers": "authorization,content-type",
};

const user = {
  id: 2,
  username: "figo",
  email: "figo@example.com",
  age: 25,
  role: "user",
  status: "active",
};

const admin = {
  id: 1,
  username: "admin",
  email: "admin@example.com",
  age: 30,
  role: "admin",
  status: "active",
};

const bannedUser = {
  ...user,
  status: "banned",
  ban_reason: "Policy violation",
};

type MockProfile = "anonymous" | "user" | "admin" | "banned";

async function fulfillJson(route: Route, data: unknown, status = 200) {
  if (route.request().method() === "OPTIONS") {
    await route.fulfill({ status: 204, headers: corsHeaders });
    return;
  }

  await route.fulfill({
    status,
    headers: {
      ...corsHeaders,
      "content-type": "application/json",
    },
    body: JSON.stringify(data),
  });
}

async function seedSession(page: Page, profile: Exclude<MockProfile, "anonymous">) {
  const sessionUser = profile === "admin" ? admin : profile === "banned" ? bannedUser : user;
  await page.addInitScript(
    ({ sessionUser }) => {
      window.localStorage.setItem(
        "mygram-auth",
        JSON.stringify({
          state: {
            token: `${sessionUser.role}-token`,
            user: sessionUser,
            notice: null,
            isAuthenticated: true,
          },
          version: 0,
        }),
      );
    },
    { sessionUser },
  );
}

async function mockApi(page: Page, profile: MockProfile = "anonymous") {
  let activeUser = profile === "admin" ? admin : profile === "banned" ? bannedUser : user;
  const photos = [
    {
      id: 1,
      title: "Existing photo",
      caption: "Already in the feed",
      photo_url: "https://images.unsplash.com/photo-1500530855697-b586d89ba3ee",
      user_id: 2,
      created_at: new Date().toISOString(),
    },
  ];
  const comments = [
    {
      id: 1,
      message: "Existing comment",
      photo_id: 1,
      user_id: 2,
      created_at: new Date().toISOString(),
    },
  ];
  const socialLinks = [
    {
      id: 1,
      name: "GitHub",
      social_media_url: "https://github.com/example",
      user_id: 2,
    },
  ];

  await page.route("**/openapi/public.json", (route) =>
    fulfillJson(route, {
      swagger: "2.0",
      info: { title: "MyGram Public API" },
      paths: {
        "/api/v1/photos": {},
        "/api/v1/uploads/photos": {},
        "/api/v1/photos/{photoId}/comments": {},
      },
    }),
  );

  await page.route("**/api/v1/auth/register", (route) => fulfillJson(route, user, 201));

  await page.route("**/api/v1/auth/login", (route) =>
    fulfillJson(route, {
      token: "user-token",
      user,
    }),
  );

  await page.route("**/api/v1/me", (route) => {
    if (profile === "banned") {
      return fulfillJson(route, { message: "Account is banned" }, 403);
    }

    if (route.request().method() === "PATCH") {
      activeUser = {
        ...activeUser,
        ...route.request().postDataJSON(),
      };
      return fulfillJson(route, { user: activeUser });
    }

    return fulfillJson(route, {
      user: activeUser,
    });
  });

  await page.route("**/api/v1/photos", async (route) => {
    if (route.request().method() === "POST") {
      const payload = route.request().postDataJSON();
      const nextPhoto = {
        id: photos.length + 1,
        ...payload,
        user_id: 2,
        created_at: new Date().toISOString(),
      };
      photos.unshift(nextPhoto);
      await fulfillJson(route, nextPhoto, 201);
      return;
    }

    await fulfillJson(route, photos);
  });

  await page.route("**/api/v1/photos/1", (route) => fulfillJson(route, photos[0]));

  await page.route("**/api/v1/uploads/photos", (route) =>
    fulfillJson(
      route,
      {
        url: "https://s3.fgdev.tech/fgdev-media/uploads/photos/2/smoke.png",
        key: "uploads/photos/2/smoke.png",
        bucket: "fgdev-media",
        content_type: "image/png",
        size: 67,
      },
      201,
    ),
  );

  await page.route("**/api/v1/photos/1/comments", async (route) => {
    if (route.request().method() === "POST") {
      const payload = route.request().postDataJSON();
      const nextComment = {
        id: comments.length + 1,
        ...payload,
        photo_id: 1,
        user_id: 2,
        created_at: new Date().toISOString(),
      };
      comments.push(nextComment);
      await fulfillJson(route, nextComment, 201);
      return;
    }

    await fulfillJson(route, comments);
  });

  await page.route("**/api/v1/social-media", async (route) => {
    if (route.request().method() === "POST") {
      const payload = route.request().postDataJSON();
      const nextLink = {
        id: socialLinks.length + 1,
        ...payload,
        user_id: 2,
      };
      socialLinks.unshift(nextLink);
      await fulfillJson(route, nextLink, 201);
      return;
    }

    await fulfillJson(route, socialLinks);
  });

  await page.route("**/api/v1/admin/stats", (route) =>
    fulfillJson(route, {
      total_users: 2,
      active_users: 1,
      banned_users: 1,
      admin_users: 1,
      users_seen_last_24h: 1,
      total_photos: photos.length,
      total_comments: comments.length,
      total_social_media: socialLinks.length,
      recent_users: [admin, user],
      generated_at: new Date().toISOString(),
    }),
  );

  await page.route("**/api/v1/admin/users**", (route) =>
    fulfillJson(route, {
      users: [admin, user],
      total: 2,
      page: 1,
      limit: 10,
    }),
  );
}

test("registers and logs in", async ({ page }) => {
  await mockApi(page);

  await page.goto("/register");
  await page.getByLabel("Username").fill("figo");
  await page.getByLabel("Email").fill("figo@example.com");
  await page.getByLabel("Age").fill("25");
  await page.getByLabel("Password").fill("secret123");
  await page.getByRole("button", { name: "Create account" }).click();
  await expect(page.getByText("Login to MyGram")).toBeVisible();

  await page.getByLabel("Email").fill("figo@example.com");
  await page.getByLabel("Password").fill("secret123");
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page.getByRole("heading", { name: "Feed" })).toBeVisible();
});

test("creates photo, comment, and social link", async ({ page }) => {
  await seedSession(page, "user");
  await mockApi(page, "user");

  await page.goto("/feed");
  await page.getByLabel("Title").fill("Smoke photo");
  await page.getByLabel("Image file").setInputFiles({
    name: "smoke.png",
    mimeType: "image/png",
    buffer: Buffer.from([
      0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
      0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
    ]),
  });
  await page.getByLabel("Caption").fill("Created by Playwright");
  const uploadPhotoResponse = page.waitForResponse(
    (response) =>
      response.url().includes("/api/v1/uploads/photos") &&
      response.request().method() === "POST" &&
      response.status() === 201,
  );
  const createPhotoResponse = page.waitForResponse(
    (response) =>
      response.url().includes("/api/v1/photos") &&
      response.request().method() === "POST" &&
      response.status() === 201,
  );
  await page.getByRole("button", { name: "Post" }).click();
  await uploadPhotoResponse;
  await createPhotoResponse;
  await expect(page.getByText("Smoke photo")).toBeVisible({ timeout: 15_000 });
  await page.getByLabel("Search photos").fill("Smoke");
  await expect(page.getByText("Smoke photo")).toBeVisible();
  await page.getByLabel("Search photos").fill("missing");
  await expect(page.getByText("No photos match this view.")).toBeVisible();

  await page.goto("/photos/1");
  await page.getByPlaceholder("Write a comment").fill("Fresh comment");
  await page.getByRole("button", { name: "Send" }).click();
  await expect(page.getByText("Fresh comment")).toBeVisible();

  await page.goto("/social-links");
  await page.getByLabel("Name").fill("GitHub Smoke");
  await page.getByLabel("URL").fill("https://github.com/mygram-smoke");
  await page.getByRole("button", { name: "Save link" }).click();
  await expect(
    page.getByRole("link", { name: /GitHub Smoke https:\/\/github\.com\/mygram-smoke/ }),
  ).toBeVisible();

  await page.goto("/profile");
  await page.getByLabel("Username").fill("figo-updated");
  await page.getByLabel("Email").fill("figo-updated@example.com");
  await page.getByLabel("Age").fill("26");
  await page.getByRole("button", { name: "Save profile" }).click();
  await expect(page.getByRole("main").getByText("figo-updated@example.com")).toBeVisible();
});

test("serves public docs, safe try request console, and PWA manifest", async ({ page }) => {
  await mockApi(page, "user");

  await page.goto("/docs");
  await expect(page.getByRole("heading", { name: "Build with MyGram" })).toBeVisible();
  await expect(page.getByText("Try request")).toBeVisible();

  await page.getByLabel("JWT").fill("user-token");
  await page.getByRole("button", { name: "Send" }).click();
  await expect(page.getByText(/200 OK/)).toBeVisible();
  await expect(page.getByText(/figo@example\.com/)).toBeVisible();

  const manifest = await page.request.get("/manifest.webmanifest");
  expect(manifest.ok()).toBeTruthy();
  const manifestBody = await manifest.json();
  expect(manifestBody.display).toBe("standalone");
  expect(manifestBody.icons).toEqual(
    expect.arrayContaining([
      expect.objectContaining({ sizes: "192x192", type: "image/png" }),
      expect.objectContaining({ sizes: "512x512", type: "image/png" }),
    ]),
  );
});

test("enforces admin dashboard access and banned-user blocking", async ({ page }) => {
  await seedSession(page, "user");
  await mockApi(page, "user");
  await page.goto("/admin");
  await expect(page.getByRole("heading", { name: "Feed" })).toBeVisible();

  await page.evaluate(() => window.localStorage.clear());
  await seedSession(page, "admin");
  await mockApi(page, "admin");
  await page.goto("/admin");
  await expect(page.getByRole("heading", { name: "Admin Dashboard" })).toBeVisible();
  await expect(page.getByText("Total users")).toBeVisible();

  await page.evaluate(() => window.localStorage.clear());
  await seedSession(page, "banned");
  await mockApi(page, "banned");
  await page.goto("/feed");
  await expect(page.getByText("Login to MyGram")).toBeVisible();
  await expect(page.getByText("Account is banned")).toBeVisible();
});
