// weeklly, sem biblioteca (D4). Tudo aqui melhora um caminho que já funciona
// sem JavaScript: os formulários enviam e voltam; com o script, respondem no
// lugar.

const reduceMotion = matchMedia("(prefers-reduced-motion: reduce)");
const EASE = "cubic-bezier(0.2, 0.8, 0.2, 1)";

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
  if (!response.ok) throw new Error(data.error || "Não foi possível salvar. Tente de novo.");
  return data;
}

function elementFrom(html) {
  const template = document.createElement("template");
  template.innerHTML = html.trim();
  return template.content.firstElementChild;
}

// ---- Diálogos --------------------------------------------------------------
// Os links levam a uma página com o mesmo formulário; com JavaScript abrem o
// diálogo no lugar e o foco vai para o primeiro campo, ou para "Cancelar" nos
// diálogos destrutivos.

function closePopovers() {
  for (const open of document.querySelectorAll("[popover]:popover-open")) open.hidePopover();
}

for (const opener of document.querySelectorAll("[data-open-dialog]")) {
  const dialog = document.getElementById(opener.dataset.openDialog);
  if (!dialog || typeof dialog.showModal !== "function") continue;
  opener.addEventListener("click", (event) => {
    event.preventDefault();
    closePopovers();
    dialog.showModal();
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

const themeForm = document.querySelector("form.theme");
if (themeForm) {
  themeForm.addEventListener("submit", (event) => {
    event.preventDefault();
    const next = themeForm.elements.theme.value === "light" ? "light" : "dark";
    const apply = () => {
      document.documentElement.dataset.theme = next;
      themeForm.elements.theme.value = next === "light" ? "dark" : "light";
      themeForm.querySelector("button")?.setAttribute("aria-label", next === "light" ? "Mudar para o tema escuro" : "Mudar para o tema claro");
      document.querySelector('meta[name="color-scheme"]')?.setAttribute("content", next);
      document.querySelector('meta[name="theme-color"]')?.setAttribute("content", next === "light" ? "#fafafa" : "#0a0a0a");
    };
    const secure = location.protocol === "https:" ? "; Secure" : "";
    document.cookie = `weeklly_theme=${next}; Path=/; Max-Age=31536000; SameSite=Lax${secure}`;
    if (typeof document.startViewTransition === "function" && !reduceMotion.matches) {
      document.startViewTransition(apply);
    } else {
      apply();
    }
  });
}

// ---- Ordem das semanas -----------------------------------------------------
// A pílula desliza (CSS) e as listas se reordenam com animação, sem recarregar.

function byName(a, b) {
  return a.dataset.name.localeCompare(b.dataset.name, "pt-BR", { sensitivity: "base" }) || b.dataset.updated.localeCompare(a.dataset.updated);
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
    for (const button of form.querySelectorAll(".order-btn")) {
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
picker.setAttribute("aria-label", "Horários");
{
  const options = [""];
  for (let h = 0; h < 24; h++) for (const m of ["00", "30"]) options.push(`${String(h).padStart(2, "0")}:${m}`);
  picker.innerHTML = options
    .map((v) => `<button type="button" class="timepick-option${v ? "" : " is-none"}" role="option" data-value="${v}">${v || "Sem horário"}</button>`)
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
    // Sem opção exata: rola até o horário mais próximo.
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
// pede JSON e recebe a tarefa já renderizada pelo servidor.

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

if (board) {
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
      const card = form.closest(".card");
      try {
        say("Salvando…");
        const data = await send(form.action, new FormData(form));
        const item = elementFrom(data.html);
        card.querySelector(".tasks").append(item);
        refreshEmpty(card);
        form.reset();
        title.focus();
        item.scrollIntoView({ block: "nearest" });
        say("Salvo");
      } catch (error) {
        say(error.message, true);
      }
    });
  }

  // No toque não há captura de ponteiro: o clique chega ao elemento certo.
  board.addEventListener("click", (event) => focusAddFrom(event.target));

  board.addEventListener("submit", async (event) => {
    const form = event.target;
    if (!(form instanceof HTMLFormElement)) return;
    const item = form.closest(".task");
    if (!item) return;
    event.preventDefault();
    const card = item.closest(".card");
    try {
      if (form.matches(".task-toggle")) {
        const data = await send(form.action, new FormData(form));
        item.replaceWith(elementFrom(data.html));
        say("Salvo");
      } else if (form.matches(".task-remove")) {
        await send(form.action, {});
        item.remove();
        refreshEmpty(card);
        say("Tarefa excluída");
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
    const input = document.createElement("input");
    input.type = "text";
    input.className = "task-editor" + (field === "time" ? " task-editor-time" : "");
    const current = field === "time" ? item.querySelector("time")?.getAttribute("datetime") || "" : item.querySelector(".task-title").textContent;
    input.value = current;
    if (field === "time") {
      input.inputMode = "numeric";
      input.maxLength = 5;
      input.placeholder = "hh:mm";
      input.setAttribute("aria-label", "Horário");
    } else {
      input.maxLength = 200;
      input.setAttribute("aria-label", "Tarefa");
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
        say("Salvando…");
        const data = await send(`/tarefas/${item.dataset.task}/editar`, { [field]: value });
        item.replaceWith(elementFrom(data.html));
        say("Salvo");
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
      grip,
      pointerId: event.pointerId,
      dx: event.clientX - rect.left,
      dy: event.clientY - rect.top,
      width: rect.width,
      from: { list: item.parentElement, index: Array.from(item.parentElement.children).indexOf(item) },
      ghost: null,
      target: null,
    };
    grip.setPointerCapture(event.pointerId);
  });

  board.addEventListener("pointermove", (event) => {
    if (!lift || event.pointerId !== lift.pointerId) return;
    if (!lift.ghost) {
      if (Math.hypot(event.clientX - (lift.dx + lift.item.getBoundingClientRect().left), event.clientY - (lift.dy + lift.item.getBoundingClientRect().top)) < 4) return;
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
      refreshEmpty(lift.from.list.closest(".card"));
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

    const list = item.parentElement;
    const weekday = Number(list.closest(".card").dataset.weekday);
    const position = Array.from(list.children).indexOf(item);
    if (list === from.list && position === from.index) return;
    try {
      say("Salvando…");
      await send(`/tarefas/${item.dataset.task}/mover`, { weekday, position });
      say("Salvo");
    } catch (error) {
      const back = Array.from(from.list.children).filter((el) => el !== item)[from.index];
      if (back) from.list.insertBefore(item, back);
      else from.list.append(item);
      refreshEmpty(list.closest(".card"));
      refreshEmpty(from.list.closest(".card"));
      say(error.message, true);
    }
  }
  board.addEventListener("pointerup", drop);
  board.addEventListener("pointercancel", drop);

  // Teclado: Alt com seta para cima ou para baixo move a tarefa em foco.
  board.addEventListener("keydown", async (event) => {
    if (!event.altKey || (event.key !== "ArrowUp" && event.key !== "ArrowDown")) return;
    const item = event.target.closest(".task");
    if (!item) return;
    event.preventDefault();
    const list = item.parentElement;
    const index = Array.from(list.children).indexOf(item);
    const position = index + (event.key === "ArrowDown" ? 1 : -1);
    if (position < 0 || position >= list.children.length) return;
    const sibling = list.children[position];
    if (event.key === "ArrowDown") sibling.after(item);
    else sibling.before(item);
    item.querySelector(".task-title")?.focus();
    try {
      await send(`/tarefas/${item.dataset.task}/mover`, { weekday: Number(list.closest(".card").dataset.weekday), position });
      say("Salvo");
    } catch (error) {
      say(error.message, true);
    }
  });
}
