import { spawn, spawnSync } from "node:child_process";
import { existsSync, mkdirSync, readFileSync, rmSync } from "node:fs";
import { platform } from "node:os";
import { resolve } from "node:path";

const tmpPath = resolve(".tmp");
const lighthouseCli = resolve("node_modules/lighthouse/cli/index.js");
const previewBaseUrl = "http://127.0.0.1:4173";
const targets = [
  { name: "docs", url: `${previewBaseUrl}/docs` },
  { name: "login", url: `${previewBaseUrl}/login` },
];

mkdirSync(resolve("reports"), { recursive: true });
mkdirSync(tmpPath, { recursive: true });

async function isPreviewReady() {
  try {
    const response = await fetch(`${previewBaseUrl}/docs`, {
      signal: AbortSignal.timeout(2000),
    });
    return response.ok;
  } catch {
    return false;
  }
}

async function waitForPreview() {
  const startedAt = Date.now();
  while (Date.now() - startedAt < 30_000) {
    if (await isPreviewReady()) {
      return;
    }
    await new Promise((resolveWait) => setTimeout(resolveWait, 500));
  }
  throw new Error("Timed out waiting for Vite preview on port 4173");
}

let previewProcess = null;

if (!(await isPreviewReady())) {
  const isWindows = platform() === "win32";
  previewProcess = spawn(
    isWindows ? "cmd.exe" : "npm",
    isWindows
      ? ["/d", "/s", "/c", "npm.cmd run preview -- --host 127.0.0.1 --port 4173"]
      : ["run", "preview", "--", "--host", "127.0.0.1", "--port", "4173"],
    {
      cwd: process.cwd(),
      stdio: "ignore",
    },
  );
  previewProcess.unref();
  await waitForPreview();
}

function stopPreview() {
  if (!previewProcess?.pid) {
    return;
  }

  if (platform() === "win32") {
    spawnSync("taskkill", ["/pid", String(previewProcess.pid), "/t", "/f"], {
      stdio: "ignore",
    });
    return;
  }

  previewProcess.kill("SIGTERM");
}

try {
for (const target of targets) {
  const reportPath = resolve(`reports/lighthouse-${target.name}-mobile.json`);
  rmSync(reportPath, { force: true });

  const result = spawnSync(
    process.execPath,
    [
      lighthouseCli,
      target.url,
      "--chrome-flags=--headless=new --no-sandbox",
      "--only-categories=performance,accessibility,best-practices,seo",
      "--output=json",
      `--output-path=${reportPath}`,
      "--quiet",
    ],
    {
      env: {
        ...process.env,
        TEMP: tmpPath,
        TMP: tmpPath,
      },
      stdio: ["inherit", "inherit", "pipe"],
      encoding: "utf8",
    },
  );

  if (!existsSync(reportPath)) {
    if (result.stderr) {
      console.error(result.stderr);
    }
    process.exit(result.status ?? 1);
  }

  const report = JSON.parse(readFileSync(reportPath, "utf8"));

  if (report.runtimeError) {
    console.error(`Lighthouse runtime error for ${target.name}: ${report.runtimeError.code}`);
    if (result.stderr) {
      console.error(result.stderr);
    }
    process.exit(1);
  }

  const scores = {
    performance: Math.round(report.categories.performance.score * 100),
    accessibility: Math.round(report.categories.accessibility.score * 100),
    bestPractices: Math.round(report.categories["best-practices"].score * 100),
    seo: Math.round(report.categories.seo.score * 100),
  };

  console.log(
    `Lighthouse ${target.name} mobile scores: performance ${scores.performance}, accessibility ${scores.accessibility}, best-practices ${scores.bestPractices}, seo ${scores.seo}`,
  );

  const failing = Object.entries(scores).filter(([, score]) => score < 90);

  if (failing.length > 0) {
    console.error(
      `Lighthouse ${target.name} categories below 90: ${failing
        .map(([name]) => name)
        .join(", ")}`,
    );
    process.exit(1);
  }

}
} finally {
  stopPreview();
}
