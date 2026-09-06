// weeklly, sem biblioteca (D4). Tudo aqui melhora um caminho que já funciona
// sem JavaScript: os formulários enviam e voltam; com o script, respondem no
// lugar.

const reduceMotion = matchMedia("(prefers-reduced-motion: reduce)");
const EASE = "cubic-bezier(0.2, 0.8, 0.2, 1)";

// Frases do script no idioma da página (o servidor põe o dicionário no body).
const i18n = (() => {
  try {
    return JSON.parse(document.body.dataset.i18n || "{}");
  } catch {
    return {};
  }
})();
const t = (key) => i18n[key] ?? key;

// ---- Rede e avisos ---------------------------------------------------------

const status = document.querySelector("[data-status]");
let statusTimer = 0;

function say(text, isError = false) {
  if (!status) return;
  clearTimeout(statusTimer);
  status.textContent = text;
  status.classList.toggle("is-error", isError);
  if (!isError) statusTimer = setTimeout(() => { status.textContent = ""; }, 2000);
}

async function send(action, body) {
  const response = await fetch(action, {
    method: "POST",
    headers: { Accept: "application/json" },
    body: new URLSearchParams(body),
  });
  if (response.status === 204) return null;
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data.error || t("failed"));
  return data;
}

function elementFrom(html) {
  const template = document.createElement("template");
  template.innerHTML = html.trim();
  return template.content.firstElementChild;
}

// ---- Cursor ----------------------------------------------------------------
// O cursor do produto é um elemento da página que segue o ponteiro (D22), em
// vez de uma imagem em `cursor: url()`: o navegador troca imagens pelo cursor
// padrão enquanto uma página carrega e até o mouse se mexer na página nova,
// e é isso que piscava. Aqui o cursor do sistema some (html.has-cursor), o
// elemento vive no top layer para ficar acima de diálogos e menus, e a
// posição atravessa a navegação pela sessão. Sobre campos de texto e barras
// de rolagem ele some e o cursor do sistema volta. Só com ponteiro fino.

const cursor = (() => {
  const el = document.querySelector("[data-cursor]");
  if (!el || !matchMedia("(any-pointer: fine)").matches || typeof el.showPopover !== "function") return null;
  const root = document.documentElement;
  root.classList.add("has-cursor");

  let x = -100;
  let y = -100;
  let nativeTarget = false; // o alvo sob o ponteiro tem cursor do sistema
  const place = () => { el.style.transform = `translate(${x}px, ${y}px)`; };
  const raise = () => {
    try {
      if (el.matches(":popover-open")) el.hidePopover();
      el.showPopover();
    } catch { /* fora do documento */ }
  };
  raise();

  try {
    const saved = sessionStorage.getItem("weeklly_cursor");
    if (saved) {
      [x, y] = saved.split(",").map(Number);
      place();
      el.classList.add("is-visible");
    }
  } catch { /* sem sessionStorage */ }
  addEventListener("pagehide", () => {
    try { sessionStorage.setItem("weeklly_cursor", `${x},${y}`); } catch { /* idem */ }
  });

  // Barra de rolagem de um elemento (dentro da caixa, fora da área de
  // conteúdo) ou da própria página.
  function overScrollbar(event) {
    const target = event.target;
    if (event.clientX >= root.clientWidth || event.clientY >= root.clientHeight) return true;
    if (!(target instanceof Element) || target.scrollHeight <= target.clientHeight) return false;
    const rect = target.getBoundingClientRect();
    return event.clientX >= rect.left + target.clientLeft + target.clientWidth;
  }

  document.addEventListener("pointerover", (event) => {
    nativeTarget = event.target instanceof Element && getComputedStyle(event.target).cursor !== "none";
  }, { capture: true, passive: true });
  document.addEventListener("pointermove", (event) => {
    if (event.pointerType === "touch") {
      el.classList.remove("is-visible");
      return;
    }
    x = event.clientX;
    y = event.clientY;
    place();
    el.classList.toggle("is-native", nativeTarget || overScrollbar(event));
    el.classList.add("is-visible");
  }, { capture: true, passive: true });
  document.addEventListener("pointerout", (event) => {
    if (!event.relatedTarget) el.classList.remove("is-visible"); // saiu da janela
  }, { capture: true, passive: true });

  const release = () => el.classList.remove("is-pressing");
  document.addEventListener("pointerdown", (event) => {
    if (event.pointerType !== "touch" && event.button === 0) el.classList.add("is-pressing");
  }, true);
  document.addEventListener("pointerup", release, true);
  document.addEventListener("pointercancel", release, true);
  document.addEventListener("pointermove", (event) => {
    if (event.buttons === 0) release(); // soltou fora da janela
  }, { capture: true, passive: true });
  addEventListener("blur", release);

  // O que abre depois fica por cima no top layer: promover de novo.
  document.addEventListener("toggle", (event) => {
    if (event.target !== el && event.newState === "open") raise();
  }, true);

  return {
    raise,
    // Fora da captura da troca de tema: o retrato antigo sai sem o cursor e o
    // novo o mostra ao vivo, então ele não aparece congelado durante a varredura.
    pause: () => el.classList.add("is-captured"),
    resume: () => el.classList.remove("is-captured"),
  };
})();

// ---- Diálogos --------------------------------------------------------------
// Os links levam a uma página com o mesmo formulário; com JavaScript abrem o
// diálogo no lugar e o foco vai para o primeiro campo, ou para "Cancelar" nos
// diálogos destrutivos.

function closePopovers() {
  for (const open of document.querySelectorAll("[popover]:not([data-cursor]):popover-open")) open.hidePopover();
}

for (const opener of document.querySelectorAll("[data-open-dialog]")) {
  const dialog = document.getElementById(opener.dataset.openDialog);
  if (!dialog || typeof dialog.showModal !== "function") continue;
  opener.addEventListener("click", (event) => {
    event.preventDefault();
    closePopovers();
    dialog.showModal();
    cursor?.raise();
    const first = dialog.querySelector("input:not([type=hidden]), a, button");
    first?.focus();
    if (first instanceof HTMLInputElement) first.select();
  });
}
for (const closer of document.querySelectorAll("dialog [data-close-dialog]")) {
  closer.addEventListener("click", (event) => {
    event.preventDefault();
    closer.closest("dialog")?.close();
  });
}

// ---- Tema ------------------------------------------------------------------
// O cookie é o mesmo que o servidor lê: a próxima página já vem no tema certo.
// A troca usa View Transitions quando existe: o tema novo varre a tela da
// esquerda para a direita e o botão gira (animações em app.css).

function preference(name, value) {
  const secure = location.protocol === "https:" ? "; Secure" : "";
  document.cookie = `${name}=${value}; Path=/; Max-Age=31536000; SameSite=Lax${secure}`;
}

const themeForm = document.querySelector("form.theme");
if (themeForm) {
  themeForm.addEventListener("submit", (event) => {
    event.preventDefault();
    const next = themeForm.elements.theme.value === "light" ? "light" : "dark";
    const apply = () => {
      document.documentElement.dataset.theme = next;
      themeForm.elements.theme.value = next === "light" ? "dark" : "light";
      themeForm.querySelector("button")?.setAttribute("aria-label", next === "light" ? t("toDark") : t("toLight"));
      document.querySelector('meta[name="color-scheme"]')?.setAttribute("content", next);
      document.querySelector('meta[name="theme-color"]')?.setAttribute("content", next === "light" ? "#fafafa" : "#0a0a0a");
    };
    preference("weeklly_theme", next);
    if (typeof document.startViewTransition === "function" && !reduceMotion.matches) {
      cursor?.pause();
      document.startViewTransition(() => {
        apply();
        cursor?.resume();
      });
    } else {
      apply();
    }
  });
}

// ---- Idioma ----------------------------------------------------------------
// A página inteira muda de idioma, então a troca recarrega. O painel de
// idiomas é um popover aninhado no de configurações: abre ao lado dele,
// alinhado à linha "Idioma"; sem espaço ao lado (celular), abre abaixo.

for (const form of document.querySelectorAll("form[action='/idioma']")) {
  form.addEventListener("submit", (event) => {
    event.preventDefault();
    const lang = event.submitter?.value;
    if (!lang) return;
    preference("weeklly_lang", lang);
    location.reload();
  });
}
{
  const settings = document.getElementById("settings-menu");
  const panel = document.getElementById("language-menu");
  const row = settings?.querySelector("[popovertarget='language-menu']");
  if (settings && panel && row) {
    const placePanel = () => {
      const menu = settings.getBoundingClientRect();
      const anchor = row.getBoundingClientRect();
      const beside = menu.left - panel.offsetWidth - 8 >= 8;
      const top = beside ? anchor.top - 8 : menu.bottom + 8;
      panel.style.top = `${Math.max(8, Math.min(top, innerHeight - panel.offsetHeight - 8))}px`;
      panel.style.right = `${beside ? innerWidth - menu.left + 8 : innerWidth - menu.right}px`;
    };
    panel.addEventListener("toggle", (event) => {
      if (event.newState === "open") placePanel();
    });
    addEventListener("resize", () => {
      if (panel.matches(":popover-open")) placePanel();
    });
  }
}

// ---- Ordem das semanas -----------------------------------------------------
// A pílula desliza (CSS) e as listas se reordenam com animação, sem recarregar.

function byName(a, b) {
  return a.dataset.name.localeCompare(b.dataset.name, document.documentElement.lang, { sensitivity: "base" }) || b.dataset.updated.localeCompare(a.dataset.updated);
}
function byRecent(a, b) {
  return b.dataset.updated.localeCompare(a.dataset.updated);
}

function reorderList(list, order) {
  const items = Array.from(list.children);
  const before = new Map(items.map((item) => [item, item.getBoundingClientRect().top]));
  items.sort(order === "name" ? byName : byRecent);
  for (const item of items) list.append(item);
  if (reduceMotion.matches) return;
  for (const item of items) {
    const dy = before.get(item) - item.getBoundingClientRect().top;
    if (dy) item.animate([{ transform: `translateY(${dy}px)` }, { transform: "none" }], { duration: 300, easing: EASE });
  }
}

function applyOrder(order) {
  for (const form of document.querySelectorAll("form.order")) {
    form.dataset.order = order;
    form.dataset.index = order === "name" ? "1" : "0";
    for (const button of form.querySelectorAll(".segmented-btn")) {
      button.setAttribute("aria-pressed", String(button.value === order));
    }
  }
  for (const list of document.querySelectorAll("[data-week-list]")) reorderList(list, order);
}

for (const form of document.querySelectorAll("form.order")) {
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const order = event.submitter?.value;
    if (!order || order === form.dataset.order) return;
    applyOrder(order);
    try {
      await send(form.action, { order });
    } catch {
      location.reload();
    }
  });
}

// ---- Horário ---------------------------------------------------------------
// Um campo de texto que aceita "930", "9:30", "21h", mais uma lista leve de
// meia em meia hora. Vale para a nova tarefa e para editar o horário no lugar.

function parseTime(raw) {
  const text = raw.trim().toLowerCase();
  if (!text) return "";
  const match = text.match(/^(\d{1,2})(?:[:h.]?(\d{2}))?h?$/);
  if (!match) return null;
  const hour = Number(match[1]);
  const minute = match[2] ? Number(match[2]) : 0;
  if (hour > 23 || minute > 59) return null;
  return `${String(hour).padStart(2, "0")}:${String(minute).padStart(2, "0")}`;
}

const picker = document.createElement("div");
picker.className = "timepick";
picker.hidden = true;
picker.setAttribute("role", "listbox");
picker.setAttribute("aria-label", t("times"));
{
  const options = [""];
  for (let h = 0; h < 24; h++) for (const m of ["00", "30"]) options.push(`${String(h).padStart(2, "0")}:${m}`);
  picker.innerHTML = options
    .map((v) => `<button type="button" class="timepick-option${v ? "" : " is-none"}" role="option" data-value="${v}">${v || t("noTime")}</button>`)
    .join("");
}
document.body.append(picker);

let picking = null; // { input, onPick }

function highlightOption(value) {
  const wanted = parseTime(value) ?? value;
  let active = null;
  for (const option of picker.children) {
    const on = option.dataset.value === wanted;
    option.classList.toggle("is-active", on);
    if (on) active = option;
  }
  if (!active && wanted) {
    for (const option of picker.children) {
      if (option.dataset.value >= wanted) { active = option; break; }
    }
  }
  (active || picker.children[19])?.scrollIntoView({ block: "center" });
}

function placePicker(input) {
  const rect = input.getBoundingClientRect();
  const below = innerHeight - rect.bottom > 250;
  picker.style.left = `${Math.max(8, Math.min(rect.left, innerWidth - picker.offsetWidth - 8))}px`;
  picker.style.top = below ? `${rect.bottom + 6}px` : "";
  picker.style.bottom = below ? "" : `${innerHeight - rect.top + 6}px`;
}

function openPicker(input, onPick) {
  picking = { input, onPick };
  picker.hidden = false;
  placePicker(input);
  highlightOption(input.value);
}

function closePicker(commit) {
  if (!picking) return;
  const { onPick } = picking;
  picking = null;
  picker.hidden = true;
  if (commit) onPick?.();
}

picker.addEventListener("pointerdown", (event) => event.preventDefault()); // o campo não perde o foco
picker.addEventListener("click", (event) => {
  const option = event.target.closest(".timepick-option");
  if (!option || !picking) return;
  picking.input.value = option.dataset.value;
  closePicker(true);
});
document.addEventListener("pointerdown", (event) => {
  if (picking && !picker.contains(event.target) && event.target !== picking.input) closePicker(false);
}, true);
addEventListener("resize", () => closePicker(false));

function normalizeTime(input) {
  const value = parseTime(input.value);
  if (value !== null) input.value = value;
  return value;
}

function enhanceTimeField(input, onPick) {
  input.addEventListener("focus", () => openPicker(input, onPick));
  input.addEventListener("click", () => { if (picking?.input !== input) openPicker(input, onPick); });
  input.addEventListener("input", () => { if (picking?.input === input) highlightOption(input.value); });
  input.addEventListener("blur", () => { normalizeTime(input); closePicker(false); });
  input.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      closePicker(false);
      input.blur();
    } else if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault();
      const options = Array.from(picker.children);
      const current = options.findIndex((o) => o.classList.contains("is-active"));
      const next = options[Math.min(options.length - 1, Math.max(0, (current < 0 ? 19 : current) + (event.key === "ArrowDown" ? 1 : -1)))];
      input.value = next.dataset.value;
      highlightOption(input.value);
    } else if (event.key === "Enter") {
      normalizeTime(input);
      closePicker(false);
    }
  });
}

// ---- Faixa do quadro -------------------------------------------------------
// - abre com o cartão de hoje no centro;
// - roda vertical do mouse desliza a faixa na horizontal;
// - arrastar com o mouse navega e solta com inércia; parado, para onde ficou;
// - setas do teclado, "Ir para hoje" e o paginador centralizam um cartão;
// - o paginador acende os dias que estão na tela.

const board = document.querySelector(".board");
const cards = board ? Array.from(board.querySelectorAll(".card")) : [];
const dots = Array.from(document.querySelectorAll(".pager-dot"));

let inertia = 0;

function stopInertia() {
  cancelAnimationFrame(inertia);
  inertia = 0;
}

function centerOn(card, smooth) {
  stopInertia();
  const left = card.offsetLeft - (board.clientWidth - card.offsetWidth) / 2;
  board.scrollTo({ left, behavior: smooth && !reduceMotion.matches ? "smooth" : "auto" });
}

function nearestIndex() {
  const center = board.scrollLeft + board.clientWidth / 2;
  let best = 0;
  let bestDistance = Infinity;
  cards.forEach((card, i) => {
    const distance = Math.abs(card.offsetLeft + card.offsetWidth / 2 - center);
    if (distance < bestDistance) {
      bestDistance = distance;
      best = i;
    }
  });
  return best;
}

function refreshEmpty(card) {
  const empty = card.querySelector(".tasks-empty");
  if (empty) empty.hidden = Boolean(card.querySelector(".task"));
}

// Clicar no espaço vazio de um dia começa uma tarefa nova ali.
function focusAddFrom(target) {
  if (!(target instanceof Element) || target.closest(".task, button, a, input, form")) return;
  const card = target.closest(".card");
  if (!card || !target.closest(".card-body")) return;
  card.querySelector(".task-add-title")?.focus();
}

if (board && cards.length) {
  const today = board.querySelector(".card.is-today");
  if (today) centerOn(today, false);

  for (const link of document.querySelectorAll('[data-action="today"]')) {
    link.addEventListener("click", (event) => {
      if (!today) return;
      event.preventDefault();
      centerOn(today, true);
    });
  }

  for (const dot of dots) {
    dot.addEventListener("click", () => {
      const card = cards[Number(dot.dataset.weekday)];
      if (card) centerOn(card, true);
    });
  }

  const visible = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        const dot = dots[cards.indexOf(entry.target)];
        if (dot) dot.classList.toggle("is-visible", entry.intersectionRatio >= 0.5);
      }
    },
    { root: board, threshold: [0.5] },
  );
  for (const card of cards) visible.observe(card);

  board.addEventListener("keydown", (event) => {
    if (event.target !== board) return; // dentro de um campo, as setas são do campo
    const step = { ArrowLeft: -1, ArrowRight: 1, Home: -Infinity, End: Infinity }[event.key];
    if (step === undefined) return;
    event.preventDefault();
    const index = Math.min(cards.length - 1, Math.max(0, nearestIndex() + step));
    centerOn(cards[index], true);
  });

  board.addEventListener(
    "wheel",
    (event) => {
      const body = event.target.closest(".card-body");
      if (body && body.scrollHeight > body.clientHeight) return; // a lista do dia rola sozinha
      const unit = event.deltaMode === 1 ? 16 : event.deltaMode === 2 ? board.clientWidth : 1;
      const dx = event.deltaX * unit;
      const dy = event.deltaY * unit;
      if (Math.abs(dy) <= Math.abs(dx)) return;
      stopInertia();
      board.scrollLeft += dy;
      event.preventDefault();
    },
    { passive: false },
  );
  board.addEventListener("scroll", () => closePicker(false), { passive: true });

  let panning = null;
  let swallowClick = false;

  board.addEventListener("pointerdown", (event) => {
    if (event.pointerType !== "mouse" || event.button !== 0) return;
    if (event.target.closest("a, button, input, textarea, select, form, [contenteditable]")) return;
    stopInertia();
    panning = { x: event.clientX, scroll: board.scrollLeft, velocity: 0, at: performance.now(), moved: false, target: event.target };
    board.setPointerCapture(event.pointerId);
    board.classList.add("is-dragging");
  });

  board.addEventListener("pointermove", (event) => {
    if (!panning) return;
    const now = performance.now();
    const target = panning.scroll - (event.clientX - panning.x);
    const elapsed = now - panning.at;
    if (elapsed > 0) panning.velocity = (target - board.scrollLeft) / elapsed;
    board.scrollLeft = target;
    panning.at = now;
    if (Math.abs(event.clientX - panning.x) > 4) panning.moved = true;
  });

  const release = () => {
    if (!panning) return;
    const { velocity, at, moved, target } = panning;
    panning = null;
    board.classList.remove("is-dragging");
    if (moved) swallowClick = true;
    // Com a captura do ponteiro, o clique chega à faixa, não ao cartão: um
    // toque parado no espaço vazio de um dia é tratado aqui.
    if (!moved) focusAddFrom(target);

    const stale = performance.now() - at > 80;
    if (stale || reduceMotion.matches || Math.abs(velocity) < 0.05) return;

    let speed = velocity;
    let last = performance.now();
    const glide = (now) => {
      const elapsed = now - last;
      last = now;
      board.scrollLeft += speed * elapsed;
      speed *= Math.pow(0.996, elapsed);
      inertia = Math.abs(speed) > 0.02 ? requestAnimationFrame(glide) : 0;
    };
    inertia = requestAnimationFrame(glide);
  };
  board.addEventListener("pointerup", release);
  board.addEventListener("pointercancel", release);

  board.addEventListener(
    "click",
    (event) => {
      if (!swallowClick) return;
      swallowClick = false;
      event.preventDefault();
      event.stopPropagation();
    },
    true,
  );

  let resizing = 0;
  let anchored = nearestIndex();
  board.addEventListener("scroll", () => {
    if (!resizing) anchored = nearestIndex();
  }, { passive: true });
  addEventListener("resize", () => {
    clearTimeout(resizing);
    resizing = setTimeout(() => {
      resizing = 0;
    }, 200);
    centerOn(cards[anchored], false);
  });
}

// ---- Tarefas ---------------------------------------------------------------
// Cada ação é um formulário que funciona sem script. Com script, o mesmo POST
// pede JSON e recebe a tarefa já renderizada pelo servidor. Cada ação entra
// no histórico com sua inversa, para desfazer e refazer.

if (board) {
  const weekID = board.dataset.week;

  // Leitura e escrita do DOM das tarefas.
  const listOf = (weekday) => board.querySelector(`.card[data-weekday="${weekday}"] .tasks`);
  const findTask = (id) => board.querySelector(`.task[data-task="${id}"]`);
  const positionOf = (item) => Array.from(item.parentElement.children).indexOf(item);
  const weekdayOf = (item) => Number(item.closest(".card").dataset.weekday);
  const snapshot = (item) => ({
    id: item.dataset.task,
    title: item.querySelector(".task-title").textContent,
    time: item.querySelector("time")?.getAttribute("datetime") || "",
    done: item.classList.contains("is-done"),
    weekday: weekdayOf(item),
    position: positionOf(item),
  });

  function placeTask(html, weekday, position) {
    const item = elementFrom(html);
    const list = listOf(weekday);
    const siblings = Array.from(list.children);
    if (position >= siblings.length) list.append(item);
    else list.insertBefore(item, siblings[position]);
    refreshEmpty(list.closest(".card"));
    return item;
  }
  function replaceTask(id, html) {
    const item = findTask(id);
    const fresh = elementFrom(html);
    item?.replaceWith(fresh);
    return fresh;
  }
  function removeTask(id) {
    const item = findTask(id);
    if (!item) return;
    const card = item.closest(".card");
    item.remove();
    refreshEmpty(card);
  }
  function moveTaskInDOM(id, weekday, position) {
    const item = findTask(id);
    if (!item) return;
    const from = item.closest(".card");
    const list = listOf(weekday);
    const siblings = Array.from(list.children).filter((el) => el !== item);
    if (position >= siblings.length) list.append(item);
    else list.insertBefore(item, siblings[position]);
    refreshEmpty(from);
    refreshEmpty(list.closest(".card"));
  }

  // Chamadas ao servidor, uma por ação.
  const api = {
    add: (weekday, title, time) => send(`/semana/${weekID}/tarefas`, { weekday, title, time }),
    edit: (id, fields) => send(`/tarefas/${id}/editar`, fields),
    done: (id, done) => send(`/tarefas/${id}/concluir`, { done: done ? "1" : "0" }),
    move: (id, weekday, position) => send(`/tarefas/${id}/mover`, { weekday, position }),
    remove: (id) => send(`/tarefas/${id}/excluir`, {}),
  };

  // Recria uma tarefa apagada com o mesmo conteúdo, estado e lugar.
  async function restore(snap) {
    const data = await api.add(snap.weekday, snap.title, snap.time);
    let html = data.html;
    if (snap.done) html = (await api.done(data.id, true)).html;
    const total = listOf(snap.weekday).children.length;
    if (snap.position < total) html = (await api.move(data.id, snap.weekday, snap.position)).html;
    placeTask(html, snap.weekday, snap.position);
    return data.id;
  }

  // ---- Histórico -----------------------------------------------------------
  const undoStack = [];
  const redoStack = [];
  const undoButton = document.querySelector("[data-undo]");
  const redoButton = document.querySelector("[data-redo]");
  let busy = false;

  function refreshHistory() {
    if (undoButton) undoButton.disabled = busy || undoStack.length === 0;
    if (redoButton) redoButton.disabled = busy || redoStack.length === 0;
  }
  function record(entry) {
    undoStack.push(entry);
    redoStack.length = 0;
    refreshHistory();
  }
  async function step(from, to, label) {
    const entry = from.pop();
    if (!entry || busy) return;
    busy = true;
    refreshHistory();
    try {
      await (from === undoStack ? entry.undo() : entry.redo());
      to.push(entry);
      say(label);
    } catch (error) {
      say(error.message, true);
    } finally {
      busy = false;
      refreshHistory();
    }
  }
  undoButton?.addEventListener("click", () => step(undoStack, redoStack, t("undone")));
  redoButton?.addEventListener("click", () => step(redoStack, undoStack, t("redone")));
  document.addEventListener("keydown", (event) => {
    if (!(event.ctrlKey || event.metaKey) || event.altKey) return;
    if (event.target.closest("input, textarea, [contenteditable]")) return;
    const key = event.key.toLowerCase();
    if (key === "z" && !event.shiftKey) {
      event.preventDefault();
      step(undoStack, redoStack, t("undone"));
    } else if ((key === "z" && event.shiftKey) || key === "y") {
      event.preventDefault();
      step(redoStack, undoStack, t("redone"));
    }
  });
  refreshHistory();

  // ---- Ações ---------------------------------------------------------------

  for (const form of board.querySelectorAll(".task-add")) {
    enhanceTimeField(form.elements.time);
    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      const title = form.elements.title;
      if (!title.value.trim()) {
        title.focus();
        return;
      }
      normalizeTime(form.elements.time);
      const weekday = Number(form.dataset.weekday);
      const values = { title: title.value, time: form.elements.time.value };
      try {
        say(t("saving"));
        const data = await api.add(weekday, values.title, values.time);
        const item = placeTask(data.html, weekday, Infinity);
        form.reset();
        title.focus();
        item.scrollIntoView({ block: "nearest" });
        say(t("saved"));
        const snap = snapshot(item);
        record({
          undo: async () => { await api.remove(snap.id); removeTask(snap.id); },
          redo: async () => { snap.id = await restore(snap); },
        });
      } catch (error) {
        say(error.message, true);
      }
    });
  }

  board.addEventListener("click", (event) => focusAddFrom(event.target));

  board.addEventListener("submit", async (event) => {
    const form = event.target;
    if (!(form instanceof HTMLFormElement)) return;
    const item = form.closest(".task");
    if (!item) return;
    event.preventDefault();
    const snap = snapshot(item);
    try {
      if (form.matches(".task-toggle")) {
        const done = !snap.done;
        replaceTask(snap.id, (await api.done(snap.id, done)).html);
        say(t("saved"));
        record({
          undo: async () => replaceTask(snap.id, (await api.done(snap.id, !done)).html),
          redo: async () => replaceTask(snap.id, (await api.done(snap.id, done)).html),
        });
      } else if (form.matches(".task-remove")) {
        await api.remove(snap.id);
        removeTask(snap.id);
        say(t("deleted"));
        record({
          undo: async () => { snap.id = await restore(snap); },
          redo: async () => { await api.remove(snap.id); removeTask(snap.id); },
        });
      }
    } catch (error) {
      say(error.message, true);
    }
  });

  // Edição no lugar: o título e o horário são botões; clicar troca por um
  // campo. Enter ou sair do campo salva; Esc desiste.
  board.addEventListener("click", (event) => {
    const button = event.target.closest("[data-edit]");
    if (!button) return;
    const item = button.closest(".task");
    const field = button.dataset.edit;
    const before = snapshot(item);
    const input = document.createElement("input");
    input.type = "text";
    input.className = "task-editor" + (field === "time" ? " task-editor-time" : "");
    const current = field === "time" ? before.time : before.title;
    input.value = current;
    if (field === "time") {
      input.inputMode = "numeric";
      input.maxLength = 5;
      input.placeholder = "hh:mm";
      input.setAttribute("aria-label", t("time"));
    } else {
      input.maxLength = 200;
      input.setAttribute("aria-label", t("task"));
    }
    button.replaceWith(input);

    let finished = false;
    const finish = async (save) => {
      if (finished) return;
      finished = true;
      closePicker(false);
      let value = input.value;
      if (field === "time") value = parseTime(value) ?? value;
      const unchanged = field === "title" ? value.trim() === current.trim() : value === current;
      if (!save || unchanged || (field === "title" && !value.trim())) {
        input.replaceWith(button);
        return;
      }
      try {
        say(t("saving"));
        replaceTask(before.id, (await api.edit(before.id, { [field]: value })).html);
        say(t("saved"));
        record({
          undo: async () => replaceTask(before.id, (await api.edit(before.id, { [field]: current })).html),
          redo: async () => replaceTask(before.id, (await api.edit(before.id, { [field]: value })).html),
        });
      } catch (error) {
        input.replaceWith(button);
        say(error.message, true);
      }
    };
    if (field === "time") enhanceTimeField(input, () => finish(true));
    input.addEventListener("keydown", (keyEvent) => {
      if (keyEvent.key === "Enter") {
        keyEvent.preventDefault();
        finish(true);
      } else if (keyEvent.key === "Escape") {
        keyEvent.preventDefault();
        finish(false);
      }
    });
    input.addEventListener("blur", () => finish(true));
    input.focus();
    if (field === "title") input.select();
  });

  // Mover, com histórico. Usado pelo arrasto e pelo teclado.
  async function commitMove(item, from) {
    const to = { weekday: weekdayOf(item), position: positionOf(item) };
    if (to.weekday === from.weekday && to.position === from.position) return;
    try {
      say(t("saving"));
      await api.move(from.id, to.weekday, to.position);
      say(t("saved"));
      record({
        undo: async () => { await api.move(from.id, from.weekday, from.position); moveTaskInDOM(from.id, from.weekday, from.position); },
        redo: async () => { await api.move(from.id, to.weekday, to.position); moveTaskInDOM(from.id, to.weekday, to.position); },
      });
    } catch (error) {
      moveTaskInDOM(from.id, from.weekday, from.position);
      say(error.message, true);
    }
  }

  // ---- Reordenar arrastando ------------------------------------------------
  // Pela alça: o fantasma segue o ponteiro, a tarefa (apagada) mostra o
  // lugar ao vivo; soltar salva. Serve para trocar de dia também.

  let lift = null;

  function listAt(x, y) {
    const card = document.elementFromPoint(x, y)?.closest(".card");
    return card?.querySelector(".tasks") ?? null;
  }

  function placeIn(list, item, y) {
    const siblings = Array.from(list.children).filter((el) => el !== item);
    const next = siblings.find((el) => {
      const rect = el.getBoundingClientRect();
      return y < rect.top + rect.height / 2;
    });
    if (next) list.insertBefore(item, next);
    else list.append(item);
  }

  board.addEventListener("pointerdown", (event) => {
    const grip = event.target.closest(".task-grip");
    if (!grip || event.button !== 0) return;
    event.preventDefault();
    const item = grip.closest(".task");
    const rect = item.getBoundingClientRect();
    lift = {
      item,
      pointerId: event.pointerId,
      dx: event.clientX - rect.left,
      dy: event.clientY - rect.top,
      width: rect.width,
      from: snapshot(item),
      ghost: null,
      target: null,
    };
    grip.setPointerCapture(event.pointerId);
  });

  board.addEventListener("pointermove", (event) => {
    if (!lift || event.pointerId !== lift.pointerId) return;
    if (!lift.ghost) {
      const rect = lift.item.getBoundingClientRect();
      if (Math.hypot(event.clientX - (lift.dx + rect.left), event.clientY - (lift.dy + rect.top)) < 4) return;
      lift.ghost = lift.item.cloneNode(true);
      lift.ghost.classList.add("task-ghost");
      lift.ghost.style.width = `${lift.width}px`;
      document.body.append(lift.ghost);
      lift.item.classList.add("is-placeholder");
      closePicker(false);
    }
    lift.ghost.style.left = `${event.clientX - lift.dx}px`;
    lift.ghost.style.top = `${event.clientY - lift.dy}px`;

    const list = listAt(event.clientX, event.clientY);
    if (list) {
      if (lift.target && lift.target !== list) lift.target.closest(".card").classList.remove("is-drop-target");
      lift.target = list;
      list.closest(".card").classList.add("is-drop-target");
      placeIn(list, lift.item, event.clientY);
      refreshEmpty(list.closest(".card"));
      refreshEmpty(listOf(lift.from.weekday).closest(".card"));
    }
  });

  async function drop(event) {
    if (!lift || event.pointerId !== lift.pointerId) return;
    const { item, ghost, from, target } = lift;
    lift = null;
    if (!ghost) return;
    ghost.remove();
    item.classList.remove("is-placeholder");
    target?.closest(".card").classList.remove("is-drop-target");
    await commitMove(item, from);
  }
  board.addEventListener("pointerup", drop);
  board.addEventListener("pointercancel", drop);

  // Teclado: Alt com seta para cima ou para baixo move a tarefa em foco.
  board.addEventListener("keydown", async (event) => {
    if (!event.altKey || (event.key !== "ArrowUp" && event.key !== "ArrowDown")) return;
    const item = event.target.closest(".task");
    if (!item) return;
    event.preventDefault();
    const from = snapshot(item);
    const list = item.parentElement;
    const position = from.position + (event.key === "ArrowDown" ? 1 : -1);
    if (position < 0 || position >= list.children.length) return;
    const sibling = list.children[position];
    if (event.key === "ArrowDown") sibling.after(item);
    else sibling.before(item);
    item.querySelector(".task-title")?.focus();
    await commitMove(item, from);
  });
}
