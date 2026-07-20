import http from "k6/http";
import { check, fail, sleep } from "k6";

const apiURL = (__ENV.API_URL || "").replace(/\/$/, "");
const frontendURL = (__ENV.FRONTEND_URL || "").replace(/\/$/, "");
const profile = __ENV.PROFILE || "smoke";

const profiles = {
  smoke: [
    { duration: "10s", target: 10 },
    { duration: "20s", target: 10 },
    { duration: "10s", target: 0 },
  ],
  normal: [
    { duration: "30s", target: 100 },
    { duration: "2m", target: 100 },
    { duration: "30s", target: 0 },
  ],
  peak: [
    { duration: "30s", target: 100 },
    { duration: "30s", target: 500 },
    { duration: "1m", target: 1500 },
    { duration: "2m", target: 1500 },
    { duration: "1m", target: 0 },
  ],
};

if (!profiles[profile]) {
  throw new Error(`Unknown PROFILE ${profile}; use smoke, normal, or peak`);
}

export const options = {
  discardResponseBodies: true,
  scenarios: {
    mygram_infrastructure: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: profiles[profile],
      gracefulRampDown: "30s",
    },
  },
  thresholds: {
    checks: ["rate>0.99"],
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<1500", "p(99)<3000"],
  },
};

export function setup() {
  if (!apiURL || !frontendURL) {
    fail("API_URL and FRONTEND_URL are required");
  }

  const readiness = http.get(`${apiURL}/health/ready`, {
    responseType: "text",
    tags: { endpoint: "api-readiness-preflight" },
  });
  const ready = check(readiness, {
    "preflight status is 200": (response) => response.status === 200,
    "preflight database is connected": (response) =>
      response.json("database") === "connected",
    "preflight Redis is connected": (response) =>
      response.json("redis") === "connected",
  });
  if (!ready) {
    fail(`production readiness preflight failed: HTTP ${readiness.status}`);
  }
}

export default function () {
  const selector = (__VU + __ITER) % 10;
  let response;
  let endpoint;

  if (selector < 7) {
    endpoint = "frontend-root";
    response = http.get(`${frontendURL}/`, { tags: { endpoint } });
  } else if (selector < 9) {
    endpoint = "api-liveness";
    response = http.get(`${apiURL}/health/live`, { tags: { endpoint } });
  } else {
    endpoint = "api-readiness";
    response = http.get(`${apiURL}/health/ready`, { tags: { endpoint } });
  }

  check(response, {
    [`${endpoint} returned 200`]: (result) => result.status === 200,
  });

  // A short think time models concurrently connected users rather than a tight
  // synthetic request loop. At the 1,500 VU plateau this still creates a
  // substantial request rate across the frontend and API health paths.
  sleep(1 + Math.random() * 2);
}
