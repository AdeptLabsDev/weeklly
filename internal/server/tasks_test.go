package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// postJSON envia um formulário pedindo JSON, como o script faz.
func (a *app) postJSON(path string, form url.Values) (int, map[string]any) {
	a.t.Helper()
	req, err := http.NewRequest(http.MethodPost, a.http.URL+path, strings.NewReader(form.Encode()))
	if err != nil {
		a.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	resp, err := a.client.Do(req)
	if err != nil {
		a.t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	var data map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &data)
	}
	return resp.StatusCode, data
}

func TestTasksViaJSON(t *testing.T) {
	a := newApp(t, false)
	location := a.createWeek("Semana padrão")

	status, data := a.postJSON(location+"/tarefas", url.Values{"weekday": {"0"}, "title": {"  Treino "}, "time": {"7:00"}})
	if status != http.StatusCreated {
		t.Fatalf("criar: status %d, %v", status, data)
	}
	id, _ := data["id"].(string)
	html, _ := data["html"].(string)
	if id == "" || !strings.Contains(html, `class="task"`) || !strings.Contains(html, "Treino") || !strings.Contains(html, `datetime="07:00"`) {
		t.Errorf("resposta = %v", data)
	}

	// Validação vira 422 com frase.
	if status, data := a.postJSON(location+"/tarefas", url.Values{"weekday": {"0"}, "title": {""}}); status != http.StatusUnprocessableEntity || data["error"] == "" {
		t.Errorf("título vazio: %d %v", status, data)
	}
	if status, _ := a.postJSON(location+"/tarefas", url.Values{"weekday": {"9"}, "title": {"x"}}); status != http.StatusBadRequest {
		t.Errorf("dia inválido: %d", status)
	}
	if status, _ := a.postJSON(location+"/tarefas", url.Values{"weekday": {"0"}, "title": {"x"}, "time": {"25:00"}}); status != http.StatusUnprocessableEntity {
		t.Errorf("horário inválido: %d", status)
	}

	// Editar, concluir, desfazer.
	status, data = a.postJSON("/tarefas/"+id+"/editar", url.Values{"title": {"Treino na academia"}, "time": {""}})
	if status != http.StatusOK || !strings.Contains(data["html"].(string), "Treino na academia") || strings.Contains(data["html"].(string), "datetime=") {
		t.Errorf("editar: %d %v", status, data)
	}
	status, data = a.postJSON("/tarefas/"+id+"/concluir", url.Values{"done": {"1"}})
	if status != http.StatusOK || data["done"] != true || !strings.Contains(data["html"].(string), "is-done") {
		t.Errorf("concluir: %d %v", status, data)
	}
	status, data = a.postJSON("/tarefas/"+id+"/concluir", url.Values{"done": {"0"}})
	if status != http.StatusOK || data["done"] != false {
		t.Errorf("desfazer: %d %v", status, data)
	}
	if status, _ := a.postJSON("/tarefas/"+id+"/editar", url.Values{}); status != http.StatusBadRequest {
		t.Errorf("patch vazio: %d", status)
	}

	// O quadro mostra a tarefa e esconde o estado vazio daquele dia.
	board := a.get(location)
	if !strings.Contains(board.body, "Treino na academia") || !strings.Contains(board.body, `id="tarefa-`+id+`"`) {
		t.Error("tarefa não aparece no quadro")
	}

	// Outra pessoa não alcança a tarefa.
	stranger := newFreshClient(t, a.http)
	req, _ := http.NewRequest(http.MethodPost, a.http.URL+"/tarefas/"+id+"/excluir", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	resp, err := stranger.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("estranho apagando tarefa: %d", resp.StatusCode)
	}

	if status, _ := a.postJSON("/tarefas/"+id+"/excluir", nil); status != http.StatusNoContent {
		t.Errorf("excluir: %d", status)
	}
	if status, _ := a.postJSON("/tarefas/"+id+"/excluir", nil); status != http.StatusNotFound {
		t.Errorf("excluir de novo: %d", status)
	}
}

func TestTasksViaPlainForms(t *testing.T) {
	a := newApp(t, false)
	location := a.createWeek("Semana padrão")

	r := a.post(location+"/tarefas", url.Values{"weekday": {"5"}, "title": {"Feira"}})
	if r.status != http.StatusSeeOther || r.location != location+"#dia-5" {
		t.Fatalf("form sem script: status %d, Location %q", r.status, r.location)
	}
	board := a.get(location)
	if !strings.Contains(board.body, "Feira") {
		t.Error("tarefa criada pelo formulário não aparece")
	}
	r = a.post(location+"/tarefas", url.Values{"weekday": {"5"}, "title": {""}})
	if r.status != http.StatusUnprocessableEntity || !strings.Contains(r.body, "Não deu para salvar") {
		t.Errorf("erro sem script: status %d", r.status)
	}
}

func TestWeekActions(t *testing.T) {
	a := newApp(t, false)
	location := a.createWeek("Semana padrão")
	a.postJSON(location+"/tarefas", url.Values{"weekday": {"0"}, "title": {"Treino"}, "time": {"07:00"}})

	// Renomear: página sem script e POST.
	if r := a.get(location + "/renomear"); r.status != http.StatusOK || !strings.Contains(r.body, "Renomear semana") || !strings.Contains(r.body, `value="Semana padrão"`) {
		t.Errorf("página de renomear: %d", r.status)
	}
	if r := a.post(location+"/renomear", url.Values{"name": {"  "}}); r.status != http.StatusUnprocessableEntity {
		t.Errorf("renomear vazio: %d", r.status)
	}
	if r := a.post(location+"/renomear", url.Values{"name": {"Semana base"}}); r.status != http.StatusSeeOther || r.location != location {
		t.Errorf("renomear: %d %q", r.status, r.location)
	}
	if b := a.get(location); !strings.Contains(b.body, `<span class="nav-week-name">Semana base</span>`) {
		t.Error("nome novo não apareceu na barra")
	}

	// Duplicar leva para a cópia, com as tarefas.
	dup := a.post(location+"/duplicar", nil)
	if dup.status != http.StatusSeeOther || !strings.HasPrefix(dup.location, "/semana/") || dup.location == location {
		t.Fatalf("duplicar: %d %q", dup.status, dup.location)
	}
	copyBoard := a.get(dup.location)
	if !strings.Contains(copyBoard.body, "Semana base (cópia)") || !strings.Contains(copyBoard.body, "Treino") || !strings.Contains(copyBoard.body, `datetime="07:00"`) {
		t.Error("cópia sem nome ou sem tarefas")
	}

	// Excluir: página de confirmação e POST; depois some.
	if r := a.get(dup.location + "/excluir"); r.status != http.StatusOK || !strings.Contains(r.body, "Excluir Semana base (cópia)?") {
		t.Errorf("página de excluir: %d", r.status)
	}
	if r := a.post(dup.location+"/excluir", nil); r.status != http.StatusSeeOther || r.location != "/semanas" {
		t.Errorf("excluir: %d %q", r.status, r.location)
	}
	if r := a.get(dup.location); r.status != http.StatusNotFound {
		t.Errorf("cópia ainda existe: %d", r.status)
	}

	// Ordem: A–Z e recentes, lembrada por usuário.
	a.createWeek("Ano novo")
	hub := a.get("/semanas")
	if strings.Index(hub.body, "Ano novo") > strings.Index(hub.body, "Semana base") {
		t.Error("recentes: a semana nova deveria vir primeiro")
	}
	if r := a.post("/semanas/ordem", url.Values{"order": {"name"}}); r.status != http.StatusSeeOther {
		t.Errorf("ordem: %d", r.status)
	}
	hub = a.get("/semanas")
	if strings.Index(hub.body, "Ano novo") > strings.Index(hub.body, "Semana base") || !strings.Contains(hub.body, `data-order="name"`) {
		t.Error("A–Z: a ordem não mudou ou o controle não reflete")
	}
	if status, _ := a.postJSON("/semanas/ordem", url.Values{"order": {"recent"}}); status != http.StatusNoContent {
		t.Errorf("ordem via script: %d", status)
	}

	// Mover tarefa: dentro do dia e para outro dia.
	_, second := a.postJSON(location+"/tarefas", url.Values{"weekday": {"0"}, "title": {"Leitura"}})
	id, _ := second["id"].(string)
	status, data := a.postJSON("/tarefas/"+id+"/mover", url.Values{"weekday": {"0"}, "position": {"0"}})
	if status != http.StatusOK || data["position"] != float64(0) {
		t.Errorf("mover para o início: %d %v", status, data)
	}
	status, data = a.postJSON("/tarefas/"+id+"/mover", url.Values{"weekday": {"3"}, "position": {"5"}})
	if status != http.StatusOK || data["weekday"] != float64(3) || data["position"] != float64(0) {
		t.Errorf("mover para quinta: %d %v", status, data)
	}
	if status, _ := a.postJSON("/tarefas/"+id+"/mover", url.Values{"weekday": {"8"}, "position": {"0"}}); status != http.StatusBadRequest {
		t.Errorf("dia inválido: %d", status)
	}
	if status, _ := a.postJSON("/tarefas/"+id+"/mover", url.Values{"weekday": {"0"}, "position": {"-1"}}); status != http.StatusBadRequest {
		t.Errorf("posição inválida: %d", status)
	}
	board := a.get(location)
	if !strings.Contains(board.body, `id="dia-3"`) || strings.Index(board.body, "Leitura") < strings.Index(board.body, `id="dia-3"`) {
		t.Error("a tarefa não aparece na quinta")
	}
}

func TestLanguage(t *testing.T) {
	a := newApp(t, false)

	// Sem preferência: português.
	if r := a.get("/"); !strings.Contains(r.body, `lang="pt-BR"`) || !strings.Contains(r.body, "Suas semanas") {
		t.Error("padrão deveria ser português")
	}

	// Accept-Language decide a primeira visita.
	req, _ := http.NewRequest(http.MethodGet, a.http.URL+"/", nil)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,pt;q=0.8")
	resp, err := a.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	r := read(t, resp)
	if !strings.Contains(r.body, `lang="en"`) || !strings.Contains(r.body, "Your weeks") || !strings.Contains(r.body, "Sign in with Google") {
		t.Error("Accept-Language: en não aplicado")
	}

	// A escolha nas configurações vence o Accept-Language e vale para tudo.
	if r := a.post("/idioma", url.Values{"lang": {"en"}}); r.status != http.StatusSeeOther {
		t.Errorf("idioma: %d", r.status)
	}
	location := a.createWeek("Semana padrão")
	board := a.get(location)
	for _, want := range []string{`lang="en"`, "Today is Wednesday.", "Monday", "New week", "Settings", `data-i18n=`, "Saved", "Nothing here yet",
		// O seletor de idiomas lista todos, com o atual marcado.
		`id="language-menu"`, `value="en" class="menu-option" lang="en" aria-current="true"`, `value="pt-BR" class="menu-option" lang="pt-BR">`} {
		if !strings.Contains(board.body, want) {
			t.Errorf("quadro em inglês sem %q", want)
		}
	}
	if strings.Contains(board.body, "Nova semana") {
		t.Error("texto em português sobrou no quadro em inglês")
	}
	if status, data := a.postJSON(location+"/tarefas", url.Values{"weekday": {"0"}, "title": {""}}); status != http.StatusUnprocessableEntity || data["error"] != "Write what needs to be done." {
		t.Errorf("erro em inglês: %d %v", status, data)
	}
	// O apóstrofo vira &#39; no HTML; a checagem usa a frase do corpo.
	if r := a.get("/nao-existe"); !strings.Contains(r.body, "Go back to your weeks") {
		t.Error("404 em inglês")
	}

	// Volta para português; valor inválido também cai em português.
	if r := a.post("/idioma", url.Values{"lang": {"klingon"}}); r.status != http.StatusSeeOther {
		t.Errorf("idioma inválido: %d", r.status)
	}
	if r := a.get(location); !strings.Contains(r.body, "Hoje é quarta-feira.") || !strings.Contains(r.body, `lang="pt-BR" aria-current="true"`) {
		t.Error("não voltou para português")
	}
	if status, _ := a.postJSON("/idioma", url.Values{"lang": {"en"}}); status != http.StatusNoContent {
		t.Errorf("idioma via script: %d", status)
	}
}

func TestThemeCookie(t *testing.T) {
	a := newApp(t, false)
	if r := a.get("/"); !strings.Contains(r.body, `data-theme="dark"`) {
		t.Error("tema padrão deveria ser escuro")
	}
	if r := a.post("/tema", url.Values{"theme": {"light"}}); r.status != http.StatusSeeOther {
		t.Errorf("tema: %d", r.status)
	}
	r := a.get("/")
	if !strings.Contains(r.body, `data-theme="light"`) || !strings.Contains(r.body, `content="light"`) {
		t.Error("tema claro não aplicado")
	}
	if r := a.post("/tema", url.Values{"theme": {"qualquer"}}); r.status != http.StatusSeeOther {
		t.Errorf("tema inválido: %d", r.status)
	}
	if r := a.get("/"); !strings.Contains(r.body, `data-theme="dark"`) {
		t.Error("tema inválido deveria cair no escuro")
	}
}
