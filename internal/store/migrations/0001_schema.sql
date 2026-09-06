-- Modelo de dados do weeklly (roadmap §2, D12, D14, D15, D18). Tabelas
-- STRICT: o SQLite rejeita valores de tipo errado em vez de converter.
--
-- Instantes são TEXT em ISO 8601 (AAAA-MM-DDTHH:MM:SS.SSSZ), sempre em UTC,
-- para ordenar e comparar como texto sem ambiguidade. Identificadores são
-- opacos, 16 caracteres de [a-z2-7], gerados pela aplicação (ids.New).
--
-- Até o primeiro deploy este arquivo é o esquema inteiro e pode ser
-- reescrito; para recriar um banco local, apague data/. Depois do primeiro
-- deploy, toda mudança vira uma migração nova.

-- Usuários. Nascem anônimos na primeira semana criada e ganham e-mail, nome
-- e foto ao entrar com o Google (D14). Só o que o Google devolve é guardado.
-- week_order é como a pessoa prefere ver a lista de semanas.
CREATE TABLE users (
    id          TEXT PRIMARY KEY
                CHECK (length(id) = 16 AND id NOT GLOB '*[^a-z2-7]*'),
    email       TEXT UNIQUE COLLATE NOCASE,
    google_sub  TEXT UNIQUE,
    name        TEXT,
    picture_url TEXT,
    week_order  TEXT NOT NULL DEFAULT 'recent' CHECK (week_order IN ('recent', 'name')),
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
) STRICT;

-- Semanas: espaços de planejamento com nome, sem datas (D12).
CREATE TABLE weeks (
    id         TEXT PRIMARY KEY
               CHECK (length(id) = 16 AND id NOT GLOB '*[^a-z2-7]*'),
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL
               CHECK (length(name) BETWEEN 1 AND 60 AND name = trim(name)),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
) STRICT;

CREATE INDEX weeks_by_user_recent ON weeks (user_id, updated_at DESC);
CREATE INDEX weeks_by_user_name ON weeks (user_id, name COLLATE NOCASE);

-- Tarefas: os itens do plano de cada dia (D18). weekday 0 (segunda) a 6
-- (domingo); position é a ordem manual dentro do dia; time é "HH:MM" ou
-- nulo. updated_at por tarefa é a base do "último a escrever vence" (D5).
CREATE TABLE tasks (
    id         TEXT PRIMARY KEY
               CHECK (length(id) = 16 AND id NOT GLOB '*[^a-z2-7]*'),
    week_id    TEXT    NOT NULL REFERENCES weeks(id) ON DELETE CASCADE,
    weekday    INTEGER NOT NULL CHECK (weekday BETWEEN 0 AND 6),
    position   INTEGER NOT NULL CHECK (position >= 0),
    title      TEXT    NOT NULL
               CHECK (length(title) BETWEEN 1 AND 200 AND title = trim(title)),
    time       TEXT    CHECK (time IS NULL OR (time GLOB '[0-2][0-9]:[0-5][0-9]' AND time < '24:00')),
    done       INTEGER NOT NULL DEFAULT 0 CHECK (done IN (0, 1)),
    created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
) STRICT;

CREATE INDEX tasks_by_day ON tasks (week_id, weekday, position);

-- Qualquer mudança nas tarefas marca a semana como recente, para a ordem
-- do hub, e carimba a tarefa.
CREATE TRIGGER tasks_touch_week_insert AFTER INSERT ON tasks
BEGIN
    UPDATE weeks SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.week_id;
END;

CREATE TRIGGER tasks_touch_week_update AFTER UPDATE OF title, time, done, position, weekday ON tasks
BEGIN
    UPDATE tasks SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.id;
    UPDATE weeks SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = NEW.week_id;
END;

CREATE TRIGGER tasks_touch_week_delete AFTER DELETE ON tasks
BEGIN
    UPDATE weeks SET updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = OLD.week_id;
END;

-- Sessões (D15). O cookie leva o token; aqui fica só o hash dele, então um
-- vazamento do banco não entrega sessões. last_week_id lembra onde o usuário
-- estava, para abrir o app direto na semana.
CREATE TABLE sessions (
    token_hash   TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_week_id TEXT REFERENCES weeks(id) ON DELETE SET NULL,
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    last_seen_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    expires_at   TEXT NOT NULL
) STRICT;

CREATE INDEX sessions_by_user ON sessions (user_id);
CREATE INDEX sessions_by_expiry ON sessions (expires_at);
