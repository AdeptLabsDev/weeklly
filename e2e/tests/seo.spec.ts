import { test, expect } from "@playwright/test";
import fs from "node:fs";

test.beforeAll(() => fs.mkdirSync("./screenshots", { recursive: true }));

test("páginas públicas mantêm idioma e funcionam sem JavaScript @seo @shots", async ({ browser }, info) => {
  const context = await browser.newContext({
    ...info.project.use,
    javaScriptEnabled: false,
    locale: "pt-BR",
  });
  try {
    const page = await context.newPage();
    await page.goto("/en/weekly-planner");
    await expect(page.locator("html")).toHaveAttribute("lang", "en");
    await expect(page.locator("h1")).toContainText("weekly planner");
    await expect(page.locator("h1")).toHaveCount(1);
    await expect(page.locator('script:not([type="application/ld+json"])')).toHaveCount(1);
    await expect(page.locator(".landing-features li")).toHaveCount(6);
    // Abertura: a mascote ao lado do texto e o botão na cor da marca, o único da página.
    await expect(page.locator(".hero-art .mascot")).toBeVisible();
    await expect(page.locator(".landing-cta")).toHaveCSS("background-color", "rgb(114, 124, 245)");
    await expect(page.locator(".landing-header .btn-primary")).not.toHaveCSS("background-color", "rgb(114, 124, 245)");
    // A barra tem tema, idioma e o botão de abrir: nenhum atalho de seção.
    await expect(page.locator(".landing-header form.theme")).toHaveCount(1);
    await expect(page.locator(".lang-button")).toHaveText("English");
    await expect(page.locator("#language-menu")).toBeHidden();
    await expect(page.locator('.landing-header a[href^="#"]')).toHaveCount(0);
    // As perguntas saíram da landing: os pontos fortes fecham a página.
    await expect(page.locator("main > section").last()).toHaveAttribute("id", "features");
    // A faixa dos aplicativos tem fundo próprio, título na cor da marca e o selo cinza.
    await expect(page.locator(".landing-apps h2")).toHaveText("Use it anywhere, anytime");
    await expect(page.locator(".landing-apps h2")).toHaveCSS("color", "rgb(114, 124, 245)");
    expect(await page.locator(".landing-apps").evaluate((el) => getComputedStyle(el).backgroundColor)).not.toBe(await page.evaluate(() => getComputedStyle(document.body).backgroundColor));
    await expect(page.locator(".store-badge")).toContainText("Available on Google Play");
    await expect(page.locator("main details")).toHaveCount(0);
    await expect(page.locator("footer .footer-group")).toHaveCount(6);
    await expect(page.locator('meta[name="robots"]')).toHaveAttribute("content", /noindex/);
    await page.screenshot({ path: `./screenshots/seo-${info.project.name}-en.png`, fullPage: info.project.name === "desktop" });

    // Como funciona: os passos são rádios e a cena reage a eles só com CSS.
    await expect(page.locator(".how-radio")).toHaveCount(3);
    await expect(page.locator("#how-1")).toBeChecked();
    await expect(page.locator(".how-scene .how-day")).toHaveCount(7);
    await expect(page.locator(".how-scene .how-day").first()).toHaveText("mon");
    await expect(page.locator(".how-bar").first()).toHaveCSS("opacity", "0");
    await page.locator('label[for="how-2"]').click();
    await expect(page.locator("#how-2")).toBeChecked();
    await expect(page.locator(".how-bar").first()).toHaveCSS("opacity", "1");
    await expect(page.locator(".how-step").nth(1).locator(".how-body p")).toBeVisible();

    // O rodapé leva à página de perguntas, no mesmo idioma, com as respostas em <details>.
    await page.locator('footer a[href="/en/faq"]').click();
    await expect(page).toHaveURL(/\/en\/faq$/);
    await expect(page.locator("html")).toHaveAttribute("lang", "en");
    await expect(page.locator("h1")).toHaveText("Frequently asked questions");
    await expect(page.locator(".landing-questions details")).toHaveCount(8);
    const question = page.locator(".landing-questions summary").first();
    await question.focus();
    await page.keyboard.press("Enter");
    await expect(page.locator(".landing-questions details").first()).toHaveAttribute("open", "");
    await expect(page.locator(".faq-more .landing-cta")).toHaveAttribute("href", "/semanas/nova?lang=en");
    await page.screenshot({ path: `./screenshots/faq-${info.project.name}-en.png`, fullPage: info.project.name === "desktop" });
    // A tradução da página de perguntas é a página de perguntas, não a landing.
    await page.locator(".lang-button").click();
    await page.waitForTimeout(300);
    await page.locator('#language-menu a[lang="pt-BR"]').click();
    await expect(page).toHaveURL(/\/perguntas-frequentes$/);
    await expect(page.locator("h1")).toHaveText("Perguntas frequentes");
    await page.locator(".landing-header .logo").click();
    await expect(page).toHaveURL(/\/planejador-semanal$/);
    await page.goto("/en/weekly-planner");
    // O idioma abre um painel com todos os idiomas, mesmo sem script (Popover API).
    await page.locator(".lang-button").click();
    await expect(page.locator("#language-menu")).toBeVisible();
    await expect(page.locator('#language-menu a[aria-current="page"]')).toHaveText("English");
    await page.waitForTimeout(300); // entrada do painel (transição CSS) antes do clique
    await page.locator('#language-menu a[lang="pt-BR"]').click();
    await expect(page.locator("html")).toHaveAttribute("lang", "pt-BR");
    await expect(page.locator(".lang-button")).toHaveText("Português");
    await expect(page.locator("h1")).toContainText("Planejador semanal online");
    await expect(page.locator("h1")).toContainText("sem cadastro");
    // Tema sem JavaScript: o formulário troca e volta para a mesma página.
    await page.locator(".landing-header form.theme button").click();
    await expect(page).toHaveURL(/\/planejador-semanal$/);
    await expect(page.locator("html")).toHaveAttribute("data-theme", "light");
    await page.locator(".landing-header form.theme button").click();
    await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
    await page.screenshot({ path: `./screenshots/seo-${info.project.name}-pt.png`, fullPage: info.project.name === "desktop" });
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);

    // A tradução inglesa deve acompanhar a pessoa ao criar sua primeira
    // semana, mesmo em navegador português e com JavaScript desabilitado.
    await page.locator('.landing-languages a[lang="en"]').click();
    await page.locator(".landing-cta").first().click();
    await expect(page.locator("html")).toHaveAttribute("lang", "en");
    await page.locator('main input[name="name"]').fill("SEO weekly routine");
    await page.locator('main button[type="submit"]').click();
    await expect(page).toHaveURL(/\/semana\/[a-z2-7]{16}$/);
    await expect(page.locator("html")).toHaveAttribute("lang", "en");
    const board = page.url();
    await page.goto("/");
    await expect(page).toHaveURL(board);
  } finally {
    await context.close();
  }
});

test("landing cabe em 320px e carrega apenas recursos locais @seo", async ({ page }) => {
  const errors: string[] = [];
  const requests: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  page.on("request", (request) => requests.push(request.url()));
  await page.setViewportSize({ width: 320, height: 740 });
  for (const path of ["/planejador-semanal", "/en/weekly-planner", "/perguntas-frequentes", "/en/faq"]) {
    await page.goto(path);
    await expect(page.locator(".landing-cta").first()).toBeVisible();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  }
  expect(errors).toEqual([]);
  expect(requests.every((url) => new URL(url).origin === new URL(page.url()).origin)).toBe(true);
});

test("como funciona avança com a rolagem, sem prender a página, e clicar num passo leva até ele @seo", async ({ page }) => {
  await page.goto("/planejador-semanal");
  const section = page.locator("#how-it-works");
  // A seção não reserva espaço: a altura é a do conteúdo, não da tela.
  const heights = await section.evaluate((el) => [el.offsetHeight, window.innerHeight]);
  expect(heights[0]).toBeLessThan(heights[1] * 1.5);
  // Leva o centro do bloco à posição pedida na tela (0 = borda de baixo, 1 = de cima).
  const place = (at: number) => page.evaluate((at) => {
    const box = document.querySelector(".how")!.getBoundingClientRect();
    window.scrollBy(0, box.top + box.height / 2 - window.innerHeight * (1 - at));
  }, at);
  await place(0.2);
  await expect(page.locator("#how-1")).toBeChecked();
  await place(0.5);
  await expect(page.locator("#how-2")).toBeChecked();
  await expect(page.locator(".how-bar").first()).toHaveCSS("opacity", "1");
  await place(0.7);
  await expect(page.locator("#how-3")).toBeChecked();
  await expect(page.locator(".how-bar.is-move")).toHaveCSS("transform", /matrix\(1, 0, 0, 1, 144, 48\)/);
  // Clicar num passo rola a página até a posição daquele passo.
  await page.locator('label[for="how-1"]').click();
  await expect(page.locator("#how-1")).toBeChecked();
  await expect.poll(() => page.evaluate(() => {
    const box = document.querySelector(".how")!.getBoundingClientRect();
    return (window.innerHeight - (box.top + box.height / 2)) / window.innerHeight;
  })).toBeLessThan(0.38);
});
