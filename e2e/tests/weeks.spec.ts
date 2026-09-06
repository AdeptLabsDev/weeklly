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
  await expect(page.locator(`#dia-${today} .task`).first()).toHaveClass(/is-done/);

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
