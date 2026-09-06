// Gera os cards sociais a partir da marca, dos tokens e dos catálogos do app.
// Execute após `task css`: node e2e/scripts/generate-social.mjs.
import { chromium } from "@playwright/test";
import { readFile, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";

const root = new URL("../../", import.meta.url);
const read = (path) => readFile(new URL(path, root), "utf8");
const css = await read("web/static/app.css");
const font = await readFile(new URL("web/static/fonts/nunito.woff2", root));
const logo = (await read("web/templates/partials/logo.html")).match(/<svg[\s\S]*<\/svg>/)[0];
const escape = (text) => text.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]);
const browser = await chromium.launch();
try {
  const page = await browser.newPage({ viewport: { width: 1200, height: 630 }, deviceScaleFactor: 1 });
  for (const language of ["pt", "en"]) {
    const source = await read(`internal/i18n/${language}.go`);
    const catalog = Object.fromEntries([...source.matchAll(/"([\w.]+)"\s*:\s*("(?:[^"\\]|\\.)*")/g)].map((m) => [m[1], JSON.parse(m[2])]));
    const t = (key) => {
      if (!catalog[key]) throw new Error(`Tradução ausente: ${key}`);
      return escape(catalog[key]);
    };
    await page.setContent(`<!doctype html><html lang="${language === "pt" ? "pt-BR" : "en"}" data-theme="dark"><head><meta charset="utf-8"><style>${css}
      @font-face { font-family: SocialNunito; src: url(data:font/woff2;base64,${font.toString("base64")}); font-weight: 200 1000; }
      body { width: 1200px; height: 630px; padding: 48px 56px; overflow: hidden; background: var(--color-bg); color: var(--color-ink); font-family: var(--font-sans); }
      header { display: flex; align-items: center; justify-content: space-between; }
      .logo-mark { width: 164px; height: 61px; }
      header p { color: var(--color-ink-muted); font-size: 18px; }
      h1 { font: 800 64px/1.12 SocialNunito, sans-serif; letter-spacing: -.035em; margin-top: 38px; }
      .lead { margin-top: 14px; color: var(--color-ink-muted); font: 24px/1.5 var(--font-sans); }
      .days { display: grid; grid-template-columns: repeat(7, 1fr); gap: 10px; margin-top: 44px; }
      .day { height: 178px; padding: 18px 14px; border: 1px solid var(--color-border); border-radius: var(--radius-card); background: var(--color-surface); }
      h2 { font: 800 17px/1.3 SocialNunito, sans-serif; }
      .task { margin-top: 30px; display: flex; gap: 8px; align-items: center; }
      .check { flex: none; width: 12px; height: 12px; border: 1px solid var(--color-ink-muted); border-radius: 3px; }
      .line { width: 62px; height: 3px; border-radius: 3px; background: var(--color-border-strong); }
      .task + .task { margin-top: 19px; }
    </style></head><body>
      <header>${logo}<p>${t("landing.startNote")}</p></header>
      <h1>${t("seo.shareTitle")}</h1><p class="lead">${t("seo.shareBody")}</p>
      <div class="days">${Array.from({ length: 7 }, (_, i) => `<section class="day"><h2>${t(`weekday.${i}`)}</h2><div class="task"><span class="check"></span><span class="line"></span></div><div class="task"><span class="check"></span><span class="line"></span></div></section>`).join("")}</div>
    </body></html>`);
    await page.evaluate(() => document.fonts.ready);
    const image = await page.screenshot({ type: "png" });
    const destination = new URL(`web/static/og-planner-${language}.png`, root);
    await writeFile(destination, image);
    console.log(`${fileURLToPath(destination)} (${image.length} bytes)`);
  }
} finally {
  await browser.close();
}
