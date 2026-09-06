import { test, expect, type Page } from "@playwright/test";
import fs from "node:fs";

const shots = "./screenshots";
test.beforeAll(() => fs.mkdirSync(shots, { recursive: true }));

const switcher = (page: Page) => page.getByRole("button", { name: /Trocar de semana/ });

async function createWeek(page: Page, name: string) {
  const dialog = page.locator("dialog#new-week");
  await expect(dialog).toBeVisible();
  await dialog.getByLabel("Nome da semana").fill(name);
  await dialog.getByRole("button", { name: "Criar semana" }).click();
  await expect(page).toHaveURL(/\/semana\/[a-z2-7]{16}$/);
  await expect(switcher(page)).toContainText(name);
}

async function addTask(page: Page, weekday: number, title: string, time?: string) {
  const form = page.locator(`.task-add[data-weekday="${weekday}"]`);
  await form.locator('input[name="title"]').fill(title);
  if (time) await form.locator('input[name="time"]').fill(time);
  await form.locator('input[name="title"]').press("Enter");
  await expect(page.locator(`#dia-${weekday} .task`, { hasText: title })).toBeVisible();
}

test("visitante cria semanas, tarefas, troca tema e ordem @shots", async ({ page }, info) => {
  const tag = info.project.name;

  await page.goto("/");
  await expect(page.getByRole("heading", { name: "Suas semanas" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Entrar com Google" })).toBeVisible();
  await expect(switcher(page)).toHaveCount(0);
  await page.screenshot({ path: `${shots}/${tag}-01-hub-vazio.png` });

  await page.getByRole("link", { name: "Criar a primeira semana" }).click();
  await createWeek(page, "Semana padrão");
  await expect(page.locator(".card")).toHaveCount(7);
  await expect(page.locator(".card.is-today")).toHaveCount(1);
  await expect(page.locator(".tasks-empty:visible")).toHaveCount(7);

  // Clicar no espaço vazio do dia leva ao campo de nova tarefa.
  const today = Number(await page.locator(".card.is-today").getAttribute("data-weekday"));
  await page.locator(`#dia-${today} .tasks-empty`).click();
  await expect(page.locator(`.task-add[data-weekday="${today}"] input[name="title"]`)).toBeFocused();

  // Tarefas no dia de hoje; o horário aceita "7", "930" e "15:30".
  await addTask(page, today, "Treino", "7");
  await addTask(page, today, "Revisar o roadmap");
  await addTask(page, today, "Ligar para o contador", "1530");
  await expect(page.locator(`#dia-${today} .task`)).toHaveCount(3);
  await expect(page.locator(`#dia-${today} .tasks-empty`)).toBeHidden();
  await expect(page.locator(`#dia-${today} .task-time time`).first()).toHaveText("07:00");
  await expect(page.locator(`#dia-${today} .task-time time`).nth(1)).toHaveText("15:30");
  await expect(page.locator("[data-status]")).toContainText("Salvo");

  // Duplicar: a cópia entra logo abaixo, com o mesmo texto e horário; desfazer a tira.
  const treino = page.locator(`#dia-${today} .task`, { hasText: "Treino" }).first();
  await treino.hover();
  await treino.getByRole("button", { name: /Duplicar: Treino/ }).click();
  await expect(page.locator(`#dia-${today} .task`)).toHaveCount(4);
  await expect(page.locator(`#dia-${today} .task`).nth(1)).toContainText("Treino");
  await expect(page.locator(`#dia-${today} .task`).nth(1).locator("time")).toHaveText("07:00");
  await expect(page.locator("[data-status]")).toContainText("Tarefa duplicada");
  await page.getByRole("button", { name: "Desfazer" }).click();
  await expect(page.locator(`#dia-${today} .task`)).toHaveCount(3);

  // O seletor de horário abre no campo e "Sem horário" limpa.
  const timeField = page.locator(`.task-add[data-weekday="${today}"] input[name="time"]`);
  await timeField.click();
  await expect(page.locator(".timepick")).toBeVisible();
  await page.locator(".timepick-option", { hasText: "09:30" }).click();
  await expect(timeField).toHaveValue("09:30");
  await timeField.click();
  await page.locator(".timepick-option", { hasText: "Sem horário" }).click();
  await expect(timeField).toHaveValue("");
  await page.keyboard.press("Escape");

  // Concluir, editar título no lugar, editar horário, excluir.
  const first = page.locator(`#dia-${today} .task`).first();
  await first.getByRole("checkbox").click();
  // A tarefa concluída entra animando (check e linha) e termina riscada.
  await expect(page.locator(`#dia-${today} .task`).first()).toHaveClass(/is-done is-just-done/);
  await expect(page.locator(`#dia-${today} .task`).first()).not.toHaveClass(/is-just-done/);
  expect(await page.locator(`#dia-${today} .task`).first().locator(".task-title-text").evaluate((el) => getComputedStyle(el).backgroundSize)).toMatch(/^100%/);

  const second = page.locator(`#dia-${today} .task`).nth(1);
  await second.getByRole("button", { name: /Editar: Revisar o roadmap/ }).click();
  const editor = page.locator(".task-editor");
  await expect(editor).toBeFocused();
  await editor.fill("Revisar o roadmap do weeklly");
  await editor.press("Enter");
  await expect(page.locator(`#dia-${today} .task`).nth(1)).toContainText("Revisar o roadmap do weeklly");

  await page.locator(`#dia-${today} .task`).nth(1).getByRole("button", { name: "Definir horário" }).click();
  await page.locator(".task-editor-time").fill("10");
  await page.locator(".task-editor-time").press("Enter");
  await expect(page.locator(`#dia-${today} .task`).nth(1).locator("time")).toHaveText("10:00");

  // Alt+seta move pelo teclado: a terceira sobe para o meio.
  await page.locator(`#dia-${today} .task`).nth(2).locator(".task-title").focus();
  await page.keyboard.press("Alt+ArrowUp");
  await expect(page.locator(`#dia-${today} .task`).nth(1)).toContainText("Ligar para o contador");
  await expect(page.locator("[data-status]")).toContainText("Salvo");

  await page.screenshot({ path: `${shots}/${tag}-02-quadro-escuro.png` });

  await page.locator(`#dia-${today} .task`).nth(2).hover();
  await page.locator(`#dia-${today} .task`).nth(2).getByRole("button", { name: /Excluir/ }).click();
  await expect(page.locator(`#dia-${today} .task`)).toHaveCount(2);

  // As tarefas e a ordem sobrevivem a um reload: estão no servidor.
  await page.reload();
  await expect(page.locator(`#dia-${today} .task`)).toHaveCount(2);
  await expect(page.locator(`#dia-${today} .task`).first()).toHaveClass(/is-done/);
  await expect(page.locator(`#dia-${today} .task`).nth(1)).toContainText("Ligar para o contador");

  // Tema claro pelo sol/lua, sem recarregar, e lembrado no reload.
  await page.getByRole("button", { name: "Mudar para o tema claro" }).click();
  await page.waitForTimeout(220); // no meio da varredura da esquerda para a direita
  await page.screenshot({ path: `${shots}/${tag}-03a-tema-varrendo.png` });
  await expect(page.locator("html")).toHaveAttribute("data-theme", "light");
  await page.waitForTimeout(700); // a transição termina antes da captura
  await page.screenshot({ path: `${shots}/${tag}-03-quadro-claro.png` });
  await page.reload();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "light");
  await page.getByRole("button", { name: "Mudar para o tema escuro" }).click();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");

  // Segunda semana; seletor com ordem A–Z e ações da semana.
  await page.getByRole("link", { name: "Nova semana" }).first().click();
  await createWeek(page, "Ano novo");
  await switcher(page).click();
  const menu = page.locator("#weeks-menu");
  await expect(menu).toBeVisible();
  await expect(menu.locator(".menu-list .menu-item")).toHaveCount(2);
  await expect(menu.locator(".menu-list .menu-item").first()).toContainText("Ano novo"); // recentes
  await page.screenshot({ path: `${shots}/${tag}-04-seletor.png` });

  // Recentes → A–Z sem recarregar: a pílula desliza e a lista se reordena.
  // Para o teste distinguir, uma terceira semana chamada "Zebra" vai para o
  // fim em A–Z e para o começo em recentes.
  await page.keyboard.press("Escape");
  await page.getByRole("link", { name: "Nova semana" }).first().click();
  await createWeek(page, "Zebra");
  await switcher(page).click();
  await expect(menu.locator(".menu-list .menu-item").first()).toContainText("Zebra");
  await menu.getByRole("button", { name: "A–Z" }).click();
  await expect(page).toHaveURL(/\/semana\//);
  await expect(menu).toBeVisible(); // continua aberto: nada recarregou
  await expect(menu.locator("form.order")).toHaveAttribute("data-order", "name");
  await expect(menu.locator(".menu-list .menu-item").first()).toContainText("Ano novo");
  await expect(menu.locator(".menu-list .menu-item").last()).toContainText("Zebra");
  await page.reload();
  await switcher(page).click();
  await expect(page.locator("#weeks-menu form.order")).toHaveAttribute("data-order", "name");
  await expect(page.locator("#weeks-menu .menu-list .menu-item").first()).toContainText("Ano novo");
  await page.locator("#weeks-menu").getByRole("link", { name: /Semana padrão/ }).click();
  await expect(switcher(page)).toContainText("Semana padrão");

  // Renomear pelo diálogo.
  await switcher(page).click();
  await page.locator("#weeks-menu").getByRole("link", { name: "Renomear" }).click();
  const rename = page.locator("dialog#rename-week");
  await expect(rename).toBeVisible();
  await rename.getByLabel("Nome da semana").fill("Semana base");
  await rename.getByRole("button", { name: "Renomear" }).click();
  await expect(switcher(page)).toContainText("Semana base");

  // Duplicar leva as tarefas.
  await switcher(page).click();
  await page.locator("#weeks-menu").getByRole("button", { name: "Duplicar" }).click();
  await expect(switcher(page)).toContainText("Semana base (cópia)");
  await expect(page.locator(`#dia-${today} .task`)).toHaveCount(2);
  await expect(page.locator(`#dia-${today} .task.is-done`)).toHaveCount(0);

  // Excluir a cópia com confirmação.
  await switcher(page).click();
  await page.locator("#weeks-menu").getByRole("link", { name: "Excluir" }).click();
  const del = page.locator("dialog#delete-week");
  await expect(del).toBeVisible();
  await del.getByRole("button", { name: "Excluir semana" }).click();
  await expect(page).toHaveURL(/\/semanas$/);
  await expect(page.locator(".hub-item")).toHaveCount(3);
  await page.screenshot({ path: `${shots}/${tag}-05-hub.png` });

  // A logo leva ao hub.
  await page.goto(`/semana/`.replace("/semana/", "/"));
  await page.getByRole("link", { name: /weeklly, suas semanas/ }).click();
  await expect(page).toHaveURL(/\/semanas$/);

  await page.getByRole("link", { name: "Entrar com Google" }).click();
  await expect(page.getByRole("heading", { name: "Login ainda não configurado" })).toBeVisible();

  // Configurações: idioma. A linha mostra o atual; abrir lista os outros.
  // A troca recarrega a página em inglês e fica.
  await page.getByRole("button", { name: "Configurações" }).click();
  const settings = page.locator("#settings-menu");
  await expect(settings).toBeVisible();

  // Tipo de mouse: o do sistema por padrão; o do weeklly liga na hora.
  const html = page.locator("html");
  await expect(html).toHaveAttribute("data-cursor", "system");
  await expect(html).not.toHaveClass(/has-cursor/);
  await settings.getByRole("button", { name: "Mouse do weeklly" }).click();
  await expect(html).toHaveAttribute("data-cursor", "custom");
  if (info.project.name === "desktop") await expect(html).toHaveClass(/has-cursor/);

  // Cor: a laranja entra com a varredura e colore o nome e o cursor.
  await expect(html).toHaveAttribute("data-accent", "mono");
  await settings.getByRole("button", { name: "Laranja" }).click();
  await expect(html).toHaveAttribute("data-accent", "orange");
  await page.waitForTimeout(700);
  const accent = await page.locator("html").evaluate((el) => getComputedStyle(el).getPropertyValue("--color-accent").trim());
  expect(accent).toBe("#ffa057");
  await expect(settings.getByRole("button", { name: "Laranja" })).toHaveAttribute("aria-pressed", "true");

  const languageRow = settings.getByRole("button", { name: /Idioma/ });
  await expect(languageRow).toContainText("Português");
  const languages = page.locator("#language-menu");
  await expect(languages).toBeHidden();
  await languageRow.click();
  await expect(languages).toBeVisible();
  await expect(settings).toBeVisible(); // o painel abre ao lado, sem fechar o menu
  await expect(languages.getByRole("button", { name: "Português" })).toHaveAttribute("aria-current", "true");
  await page.waitForTimeout(250);
  await page.screenshot({ path: `${shots}/${tag}-06-configuracoes.png` });
  await languages.getByRole("button", { name: "English" }).click();
  await expect(page.locator("html")).toHaveAttribute("lang", "en");
  // As preferências ficam depois de recarregar.
  await expect(page.locator("html")).toHaveAttribute("data-cursor", "custom");
  await expect(page.locator("html")).toHaveAttribute("data-accent", "orange");
  await expect(page.getByRole("heading", { name: "Sign-in isn't set up yet" })).toBeVisible();
  await page.goto("/semanas");
  await expect(page.getByRole("heading", { name: "Your weeks" })).toBeVisible();
  await expect(page.getByRole("link", { name: "New week" }).first()).toBeVisible();
  await page.locator(".hub-item").first().click();
  await expect(page.locator(".subtitle")).toContainText("Today is");
  await page.screenshot({ path: `${shots}/${tag}-07-ingles.png` });
  await page.getByRole("button", { name: "Settings" }).click();
  await page.locator("#settings-menu").getByRole("button", { name: /Language/ }).click();
  await page.locator("#language-menu").getByRole("button", { name: "Português" }).click();
  await expect(page.locator("html")).toHaveAttribute("lang", "pt-BR");
});

test("desfazer e refazer cobrem adicionar, concluir, editar, mover e excluir", async ({ page }) => {
  await page.goto("/semanas/nova");
  await page.locator("main").getByLabel("Nome da semana").fill("Semana com histórico");
  await page.locator("main").getByRole("button", { name: "Criar semana" }).click();
  await expect(page).toHaveURL(/\/semana\//);

  const undo = page.getByRole("button", { name: "Desfazer" });
  const redo = page.getByRole("button", { name: "Refazer" });
  await expect(undo).toBeDisabled();
  await expect(redo).toBeDisabled();

  const today = Number(await page.locator(".card.is-today").getAttribute("data-weekday"));
  const tasks = page.locator(`#dia-${today} .task`);

  // Adicionar → desfazer some, refazer volta.
  await addTask(page, today, "Primeira", "08:00");
  await expect(undo).toBeEnabled();
  await undo.click();
  await expect(tasks).toHaveCount(0);
  await expect(page.locator("[data-status]")).toContainText("Desfeito");
  await expect(redo).toBeEnabled();
  await redo.click();
  await expect(tasks).toHaveCount(1);
  await expect(tasks.first()).toContainText("Primeira");
  await expect(tasks.first().locator("time")).toHaveText("08:00");

  // Concluir → Ctrl+Z desmarca → Ctrl+Shift+Z marca.
  await tasks.first().getByRole("checkbox").click();
  await expect(tasks.first()).toHaveClass(/is-done/);
  await page.locator("body").click({ position: { x: 5, y: 5 } });
  await page.keyboard.press("Control+z");
  await expect(tasks.first()).not.toHaveClass(/is-done/);
  await page.keyboard.press("Control+Shift+z");
  await expect(tasks.first()).toHaveClass(/is-done/);

  // Editar o título → desfazer volta ao anterior.
  await tasks.first().getByRole("button", { name: /Editar: Primeira/ }).click();
  await page.locator(".task-editor").fill("Primeira revisada");
  await page.locator(".task-editor").press("Enter");
  await expect(tasks.first()).toContainText("Primeira revisada");
  await undo.click();
  await expect(tasks.first()).toContainText("Primeira");
  await expect(tasks.first()).not.toContainText("revisada");

  // Mover pelo teclado → desfazer volta a posição.
  await addTask(page, today, "Segunda");
  await tasks.nth(1).locator(".task-title").focus();
  await page.keyboard.press("Alt+ArrowUp");
  await expect(tasks.first()).toContainText("Segunda");
  await undo.click();
  await expect(tasks.first()).toContainText("Primeira");
  await expect(tasks.nth(1)).toContainText("Segunda");

  // Excluir → desfazer recria no mesmo lugar, feita como estava.
  await tasks.first().hover();
  await tasks.first().getByRole("button", { name: /Excluir/ }).click();
  await expect(tasks).toHaveCount(1);
  await undo.click();
  await expect(tasks).toHaveCount(2);
  await expect(tasks.first()).toContainText("Primeira");
  await expect(tasks.first()).toHaveClass(/is-done/);
  await expect(tasks.first().locator("time")).toHaveText("08:00");

  // Tudo isso está no servidor.
  await page.reload();
  await expect(page.locator(`#dia-${today} .task`)).toHaveCount(2);
  await expect(page.locator(`#dia-${today} .task`).first()).toContainText("Primeira");
  await expect(page.locator(`#dia-${today} .task`).first()).toHaveClass(/is-done/);
  await expect(undo).toBeDisabled();
});

test("a faixa desliza com a roda e com o arrasto, e para onde a mão parou", async ({ page }, info) => {
  test.skip(info.project.name !== "desktop", "arrasto com mouse é coisa de desktop");

  await page.goto("/semanas/nova");
  const form = page.locator("main");
  await form.getByLabel("Nome da semana").fill("Semana de teste");
  await form.getByRole("button", { name: "Criar semana" }).click();
  await expect(page).toHaveURL(/\/semana\//);

  const board = page.locator(".board");
  const scrollLeft = () => board.evaluate((el) => el.scrollLeft);
  await expect(page.locator(".card.is-today")).toBeInViewport();

  await board.evaluate((el) => el.scrollTo({ left: 0, behavior: "instant" }));
  const start = await scrollLeft();
  await page.locator(".card").first().locator(".card-head").hover();
  await page.mouse.wheel(0, 400);
  await expect.poll(scrollLeft).toBeGreaterThan(start + 300);

  const afterWheel = await scrollLeft();
  const head = page.locator(".card").nth(1).locator(".card-head");
  const box = (await head.boundingBox())!;
  const x = box.x + box.width / 2;
  const y = box.y + box.height / 2;
  await page.mouse.move(x, y);
  await page.mouse.down();
  await page.mouse.move(x - 150, y, { steps: 10 });
  await page.waitForTimeout(150);
  await page.mouse.up();
  const afterDrag = await scrollLeft();
  expect(afterDrag).toBeGreaterThan(afterWheel + 140);
  await page.waitForTimeout(300);
  expect(await scrollLeft()).toBe(afterDrag);

  await board.focus();
  await page.keyboard.press("ArrowRight");
  await expect.poll(scrollLeft).toBeGreaterThan(afterDrag);

  // O cursor é o do produto: o do sistema some e um elemento segue o
  // ponteiro, com os arcos enquanto o botão está pressionado. Sobre campos
  // de texto ele some e o I-beam do sistema volta.
  const cursorOf = (selector: string) => page.locator(selector).first().evaluate((el) => getComputedStyle(el).cursor);
  await expect(page.locator("html")).not.toHaveClass(/has-cursor/); // desligado por padrão
  expect(await cursorOf(".card-head")).toBe("default");
  await page.context().addCookies([{ name: "weeklly_cursor", value: "custom", url: "http://127.0.0.1:8090" }]);
  await page.reload();
  await expect(page.locator("html")).toHaveClass(/has-cursor/);
  expect(await cursorOf(".card-head")).toBe("none");
  expect(await cursorOf(".btn")).toBe("none");
  expect(await cursorOf(".task-add-title")).toBe("text");
  const cursor = page.locator("[data-cursor-el]");
  await page.mouse.move(x, y);
  await expect(cursor).toHaveClass(/is-visible/);
  await expect(cursor).not.toHaveClass(/is-native/);
  expect(await cursor.evaluate((el) => el.style.transform)).toBe(`translate(${x}px, ${y}px)`);
  expect(await cursor.evaluate((el) => el.matches(":popover-open"))).toBe(true);
  await page.mouse.down();
  await expect(cursor).toHaveClass(/is-pressing/);
  await page.mouse.up();
  await expect(cursor).not.toHaveClass(/is-pressing/);
  const field = page.locator(".task-add-title").first();
  await field.hover();
  await expect(cursor).toHaveClass(/is-native/);
  await page.locator(".card-head").first().hover();
  await expect(cursor).not.toHaveClass(/is-native/);

  // A posição sobrevive à navegação: a página nova já mostra o cursor onde ele estava.
  await page.getByRole("link", { name: /weeklly, suas semanas/ }).click();
  await expect(page).toHaveURL(/\/semanas$/);
  await expect(page.locator("[data-cursor-el]")).toHaveClass(/is-visible/);
});

test("arrastar pela alça reordena dentro do dia e move para outro dia", async ({ page }, info) => {
  test.skip(info.project.name !== "desktop", "arrasto com mouse é coisa de desktop");

  await page.goto("/semanas/nova");
  await page.locator("main").getByLabel("Nome da semana").fill("Semana de arrasto");
  await page.locator("main").getByRole("button", { name: "Criar semana" }).click();
  await expect(page).toHaveURL(/\/semana\//);

  const today = Number(await page.locator(".card.is-today").getAttribute("data-weekday"));
  await addTask(page, today, "Primeira");
  await addTask(page, today, "Segunda");
  await addTask(page, today, "Terceira");
  const tasks = page.locator(`#dia-${today} .task`);

  const dragBy = async (index: number, targetY: number, targetX?: number) => {
    const grip = tasks.nth(index).locator(".task-grip");
    await tasks.nth(index).hover();
    const box = (await grip.boundingBox())!;
    await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
    await page.mouse.down();
    await page.mouse.move(box.x + box.width / 2 + 10, box.y + 10, { steps: 4 });
    await page.mouse.move(targetX ?? box.x + box.width / 2, targetY, { steps: 12 });
    await expect(page.locator(".task-ghost")).toBeVisible();
    await page.mouse.up();
    await expect(page.locator(".task-ghost")).toHaveCount(0);
    await expect(page.locator("[data-status]")).toContainText("Salvo");
  };

  // A terceira sobe para o início.
  const firstBox = (await tasks.nth(0).boundingBox())!;
  await dragBy(2, firstBox.y + 4);
  await expect(tasks.nth(0)).toContainText("Terceira");
  await expect(tasks.nth(1)).toContainText("Primeira");
  await expect(tasks.nth(2)).toContainText("Segunda");

  // "Segunda" vai para o dia vizinho que está na tela.
  const neighbour = today === 6 ? today - 1 : today + 1;
  const target = page.locator(`#dia-${neighbour} .card-body`);
  const targetBox = (await target.boundingBox())!;
  await dragBy(2, targetBox.y + targetBox.height / 2, targetBox.x + targetBox.width / 2);
  await expect(tasks).toHaveCount(2);
  await expect(page.locator(`#dia-${neighbour} .task`)).toHaveCount(1);
  await expect(page.locator(`#dia-${neighbour} .task`)).toContainText("Segunda");

  await page.reload();
  await expect(page.locator(`#dia-${today} .task`)).toHaveCount(2);
  await expect(page.locator(`#dia-${today} .task`).first()).toContainText("Terceira");
  await expect(page.locator(`#dia-${neighbour} .task`)).toContainText("Segunda");
});
