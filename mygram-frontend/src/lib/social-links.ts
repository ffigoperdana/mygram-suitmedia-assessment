const handlePattern = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/;
const youtubeChannelPattern = /^[A-Za-z0-9_-]{3,120}$/;

const blockedSegments = new Set([
  "about",
  "discover",
  "events",
  "explore",
  "groups",
  "hashtag",
  "hashtags",
  "help",
  "home",
  "intent",
  "live",
  "login",
  "marketplace",
  "p",
  "pages",
  "post",
  "posts",
  "privacy",
  "reel",
  "reels",
  "search",
  "settings",
  "share",
  "shorts",
  "signup",
  "stories",
  "story",
  "support",
  "tag",
  "tags",
  "terms",
  "videos",
  "watch",
]);

export const supportedSocialPlatforms =
  "GitHub, LinkedIn, Instagram, TikTok, YouTube, X/Twitter, Facebook, Threads, or Twitch";

export function getSocialProfileUrlError(value: string): string {
  const trimmed = value.trim();
  if (!trimmed) {
    return "Social profile URL is required.";
  }

  let parsed: URL;
  try {
    parsed = new URL(trimmed);
  } catch {
    return "Use a valid http:// or https:// URL.";
  }

  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    return "Use a valid http:// or https:// URL.";
  }

  if (parsed.port) {
    return "Social profile URL must not include a custom port.";
  }

  if (parsed.search || parsed.hash) {
    return "Use a direct profile or channel URL without search, sharing, or tracking parameters.";
  }

  const host = normalizeHost(parsed.hostname);
  const segments = getPathSegments(parsed.pathname);

  if (!segments.length || hasBlockedSegment(segments) || !isSupportedProfile(host, segments)) {
    return `Use a direct profile or channel URL from ${supportedSocialPlatforms}.`;
  }

  return "";
}

export function validateSocialProfileInput(input: HTMLInputElement): boolean {
  const message = getSocialProfileUrlError(input.value);
  input.setCustomValidity(message);
  if (message) {
    input.reportValidity();
    return false;
  }

  return true;
}

function normalizeHost(host: string) {
  return host.toLowerCase().replace(/^www\./, "").replace(/^m\./, "");
}

function getPathSegments(pathname: string) {
  try {
    return pathname
      .split("/")
      .map((segment) => decodeURIComponent(segment).trim())
      .filter(Boolean);
  } catch {
    return [];
  }
}

function hasBlockedSegment(segments: string[]) {
  return segments.some((segment) => blockedSegments.has(segment.replace(/^@/, "").toLowerCase()));
}

function isSupportedProfile(host: string, segments: string[]) {
  switch (host) {
    case "github.com":
    case "instagram.com":
    case "x.com":
    case "twitter.com":
    case "facebook.com":
    case "fb.com":
    case "twitch.tv":
      return isSingleHandle(segments);
    case "linkedin.com":
      return (
        segments.length === 2 &&
        (segments[0] === "in" || segments[0] === "company") &&
        isPlainHandle(segments[1])
      );
    case "tiktok.com":
    case "threads.net":
      return isSinglePrefixedHandle(segments, "@");
    case "youtube.com":
      return isYouTubeProfile(segments);
    default:
      return false;
  }
}

function isSingleHandle(segments: string[]) {
  return segments.length === 1 && isPlainHandle(segments[0]);
}

function isSinglePrefixedHandle(segments: string[], prefix: string) {
  return (
    segments.length === 1 &&
    segments[0].startsWith(prefix) &&
    isPlainHandle(segments[0].slice(prefix.length))
  );
}

function isPlainHandle(handle: string) {
  return handlePattern.test(handle);
}

function isYouTubeProfile(segments: string[]) {
  if (isSinglePrefixedHandle(segments, "@")) {
    return true;
  }

  if (segments.length !== 2) {
    return false;
  }

  return ["channel", "c", "user"].includes(segments[0].toLowerCase()) &&
    youtubeChannelPattern.test(segments[1]);
}
