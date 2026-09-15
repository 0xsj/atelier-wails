package desktop_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/0xsj/atelier-wails/internal/preferences/app/command"
	"github.com/0xsj/atelier-wails/internal/preferences/app/query"
	"github.com/0xsj/atelier-wails/internal/preferences/infra/memory"
	"github.com/0xsj/atelier-wails/internal/preferences/infra/storetest"
	"github.com/0xsj/atelier-wails/internal/preferences/transport/desktop"
)

const workspaceID = "01900000-0000-7000-8000-000000000001"

const createFixture = `{"ok":true,"value":{"status":"changed","entry":{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"text","text":"dark"},"revision":"1"},"event":{"name":"preference.changed","scope":{"kind":"global"},"key":"editor.theme","revision":"1","value":{"kind":"text","text":"dark"}}}}`

const conflictFixture = `{"ok":false,"failure":{"kind":"conflict","message":"preference revision conflict","type":"preferences.conflict","fields":{},"commit":"not_applied"}}`

func handler(t *testing.T, store storetest.Store) *desktop.Handler {
	t.Helper()
	commands, err := command.New(store, command.Discard{})
	if err != nil {
		t.Fatal(err)
	}
	queries, err := query.New(store)
	if err != nil {
		t.Fatal(err)
	}
	h, err := desktop.NewHandler(commands, queries)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func global() desktop.Scope { return desktop.Scope{Kind: desktop.ScopeGlobal} }

func text(value string) desktop.Value { return desktop.Value{Kind: desktop.ValueText, Text: &value} }

func intValue(value string) desktop.Value { return desktop.Value{Kind: desktop.ValueInt, Int: &value} }

func absent() desktop.Expected { return desktop.Expected{Kind: desktop.ExpectedAbsent} }

func revision(value string) desktop.Expected {
	return desktop.Expected{Kind: desktop.ExpectedRev, Revision: &value}
}

func replace(t *testing.T, h *desktop.Handler, key string, value desktop.Value, expected desktop.Expected) desktop.ReplaceOutcome {
	t.Helper()
	return h.Replace(context.Background(), desktop.ReplaceRequest{Scope: global(), Key: key, Value: value, Expected: expected})
}

func create(t *testing.T, h *desktop.Handler, key, value string) desktop.ReplaceOutcome {
	t.Helper()
	outcome := replace(t, h, key, text(value), absent())
	if !outcome.OK || outcome.Value.Status != desktop.StatusChanged {
		t.Fatalf("create %s: %+v", key, outcome.Failure)
	}
	return outcome
}

func encode(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// requireRefusal checks the failure envelope: kind, type, one field and commit.
func requireRefusal(t *testing.T, failure *desktop.Failure, kind, typ, commit string, field, problem string, secrets ...string) {
	t.Helper()
	if failure == nil {
		t.Fatal("expected failure")
	}
	if failure.Kind != kind || failure.Commit != commit {
		t.Fatalf("failure: %+v", failure)
	}
	if typ == "" && failure.Type != nil || typ != "" && (failure.Type == nil || *failure.Type != typ) {
		t.Fatalf("type: %+v", failure)
	}
	if field != "" {
		if len(failure.Fields) != 1 || failure.Fields[field] != problem {
			t.Fatalf("fields: %+v, want %s: %s", failure.Fields, field, problem)
		}
	} else if len(failure.Fields) != 0 {
		t.Fatalf("unexpected fields: %+v", failure.Fields)
	}
	encoded := encode(t, failure)
	for _, secret := range secrets {
		if strings.Contains(encoded, secret) {
			t.Fatalf("failure echoes input: %s", encoded)
		}
	}
}

func TestW01Scopes(t *testing.T) {
	h := handler(t, memory.New())
	for _, scope := range []desktop.Scope{global(), {Kind: desktop.ScopeWorkspace, WorkspaceID: workspaceID}} {
		outcome := h.Replace(context.Background(), desktop.ReplaceRequest{Scope: scope, Key: "k", Value: text("v"), Expected: absent()})
		if !outcome.OK || encode(t, outcome.Value.Entry.Scope) != encode(t, scope) {
			t.Fatalf("scope round trip: %+v", outcome)
		}
	}
	cases := []struct {
		scope          desktop.Scope
		field, problem string
	}{
		{desktop.Scope{}, "scope.kind", "missing"},
		{desktop.Scope{Kind: "tenant"}, "scope.kind", "unknown"},
		{desktop.Scope{Kind: desktop.ScopeWorkspace}, "scope.workspace_id", "missing"},
		{desktop.Scope{Kind: desktop.ScopeWorkspace, WorkspaceID: "not-a-uuid"}, "scope.workspace_id", "invalid"},
	}
	for _, c := range cases {
		outcome := h.Read(context.Background(), desktop.ReadRequest{Scope: c.scope, Key: "k"})
		requireRefusal(t, outcome.Failure, "invalid", "desktop.invalid_request", desktop.CommitNone, c.field, c.problem, "tenant", "not-a-uuid")
	}
}

func TestW02Values(t *testing.T) {
	h := handler(t, memory.New())
	flag := true
	sound := []desktop.Value{text("dark"), text(""), {Kind: desktop.ValueBool, Bool: &flag}, intValue("-9223372036854775808"), intValue("9223372036854775807"), intValue("0")}
	for i, value := range sound {
		outcome := replace(t, h, "k"+string(rune('a'+i)), value, absent())
		if !outcome.OK || encode(t, outcome.Value.Entry.Value) != encode(t, value) {
			t.Fatalf("value round trip %d: %+v", i, outcome)
		}
	}
	cases := []struct {
		value          desktop.Value
		typ            string
		field, problem string
	}{
		{intValue("abc"), "desktop.invalid_request", "value.int", "invalid"},
		{intValue("1.5"), "desktop.invalid_request", "value.int", "invalid"},
		{intValue("+1"), "desktop.invalid_request", "value.int", "invalid"},
		{intValue(""), "desktop.invalid_request", "value.int", "invalid"},
		{intValue("99999999999999999999"), "desktop.invalid_request", "value.int", "invalid"},
		{desktop.Value{Kind: desktop.ValueText}, "desktop.invalid_request", "value.text", "missing"},
		{desktop.Value{Kind: desktop.ValueBool}, "desktop.invalid_request", "value.bool", "missing"},
		{desktop.Value{Kind: desktop.ValueInt}, "desktop.invalid_request", "value.int", "missing"},
		{desktop.Value{Kind: "float"}, "desktop.invalid_request", "value.kind", "unknown"},
		{desktop.Value{}, "desktop.invalid_request", "value.kind", "missing"},
		{text(strings.Repeat("x", 4097)), "preferences.invalid_value", "", ""},
	}
	for _, c := range cases {
		outcome := replace(t, h, "zz", c.value, absent())
		requireRefusal(t, outcome.Failure, "invalid", c.typ, desktop.CommitNotApplied, c.field, c.problem, "abc", "float", "xxxx")
	}
}

func TestW03Expected(t *testing.T) {
	h := handler(t, memory.New())
	create(t, h, "editor.theme", "dark")
	same := replace(t, h, "editor.theme", text("dark"), revision("1"))
	if !same.OK || same.Value.Status != desktop.StatusSame {
		t.Fatalf("revision decode: %+v", same.Failure)
	}
	cases := []struct {
		expected       desktop.Expected
		typ            string
		field, problem string
	}{
		{revision("0"), "preferences.invalid_revision", "", ""},
		{revision("x"), "desktop.invalid_request", "expected.revision", "invalid"},
		{revision("-1"), "desktop.invalid_request", "expected.revision", "invalid"},
		{desktop.Expected{Kind: desktop.ExpectedRev}, "desktop.invalid_request", "expected.revision", "missing"},
		{desktop.Expected{Kind: "exact"}, "desktop.invalid_request", "expected.kind", "unknown"},
		{desktop.Expected{}, "desktop.invalid_request", "expected.kind", "missing"},
	}
	for _, c := range cases {
		outcome := replace(t, h, "editor.theme", text("light"), c.expected)
		requireRefusal(t, outcome.Failure, "invalid", c.typ, desktop.CommitNotApplied, c.field, c.problem, "exact", "light")
		outcome2 := h.Remove(context.Background(), desktop.RemoveRequest{Scope: global(), Key: "editor.theme", Expected: c.expected})
		requireRefusal(t, outcome2.Failure, "invalid", c.typ, desktop.CommitNotApplied, c.field, c.problem, "exact")
	}
}

func TestW04Keys(t *testing.T) {
	h := handler(t, memory.New())
	create(t, h, "editor.theme", "dark")
	for _, key := range []string{"Editor", ""} {
		outcome := h.Read(context.Background(), desktop.ReadRequest{Scope: global(), Key: key})
		requireRefusal(t, outcome.Failure, "invalid", "preferences.invalid_key", desktop.CommitNone, "", "", "Editor")
		write := replace(t, h, key, text("v"), absent())
		requireRefusal(t, write.Failure, "invalid", "preferences.invalid_key", desktop.CommitNotApplied, "", "", "Editor")
	}
}

func TestW05Read(t *testing.T) {
	h := handler(t, memory.New())
	create(t, h, "editor.theme", "dark")
	present := h.Read(context.Background(), desktop.ReadRequest{Scope: global(), Key: "editor.theme"})
	if got := encode(t, present); got != `{"ok":true,"value":{"found":true,"entry":{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"text","text":"dark"},"revision":"1"}}}` {
		t.Fatalf("present read: %s", got)
	}
	missing := h.Read(context.Background(), desktop.ReadRequest{Scope: global(), Key: "editor.font"})
	if got := encode(t, missing); got != `{"ok":true,"value":{"found":false,"entry":null}}` {
		t.Fatalf("absent read: %s", got)
	}
}

func TestW06List(t *testing.T) {
	h := handler(t, memory.New())
	create(t, h, "z.last", "z")
	create(t, h, "a.first", "a")
	listed := h.List(context.Background(), desktop.ListRequest{Scope: global()})
	if !listed.OK || len(listed.Value.Entries) != 2 || listed.Value.Entries[0].Key != "a.first" || listed.Value.Entries[1].Key != "z.last" {
		t.Fatalf("list: %+v", listed)
	}
	empty := h.List(context.Background(), desktop.ListRequest{Scope: desktop.Scope{Kind: desktop.ScopeWorkspace, WorkspaceID: workspaceID}})
	if got := encode(t, empty); got != `{"ok":true,"value":{"entries":[]}}` {
		t.Fatalf("empty list: %s", got)
	}
}

func TestW07Replace(t *testing.T) {
	h := handler(t, memory.New())
	created := create(t, h, "editor.theme", "dark")
	if created.Value.Event == nil || created.Value.Event.Revision != "1" || created.Value.Event.Value == nil {
		t.Fatalf("create event: %+v", created.Value.Event)
	}
	same := replace(t, h, "editor.theme", text("dark"), revision("1"))
	if got := encode(t, same); got != `{"ok":true,"value":{"status":"unchanged","entry":{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"text","text":"dark"},"revision":"1"},"event":null}}` {
		t.Fatalf("unchanged: %s", got)
	}
}

func TestW08Remove(t *testing.T) {
	h := handler(t, memory.New())
	create(t, h, "editor.theme", "dark")
	removed := h.Remove(context.Background(), desktop.RemoveRequest{Scope: global(), Key: "editor.theme", Expected: revision("1")})
	if got := encode(t, removed); got != `{"ok":true,"value":{"removed":true,"revision":"2","event":{"name":"preference.removed","scope":{"kind":"global"},"key":"editor.theme","revision":"2","value":null}}}` {
		t.Fatalf("removed: %s", got)
	}
	again := h.Remove(context.Background(), desktop.RemoveRequest{Scope: global(), Key: "editor.theme", Expected: absent()})
	if got := encode(t, again); got != `{"ok":true,"value":{"removed":false,"revision":null,"event":null}}` {
		t.Fatalf("absent remove: %s", got)
	}
}

func TestW09Resolve(t *testing.T) {
	h := handler(t, memory.New())
	create(t, h, "editor.theme", "dark")
	stored := h.Resolve(context.Background(), desktop.ResolveRequest{Scope: global(), Key: "editor.theme", Fallback: text("light")})
	if got := encode(t, stored); got != `{"ok":true,"value":{"value":{"kind":"text","text":"dark"},"stored":true,"revision":"1"}}` {
		t.Fatalf("stored resolve: %s", got)
	}
	flag := true
	fallback := h.Resolve(context.Background(), desktop.ResolveRequest{Scope: global(), Key: "editor.wrap", Fallback: desktop.Value{Kind: desktop.ValueBool, Bool: &flag}})
	if got := encode(t, fallback); got != `{"ok":true,"value":{"value":{"kind":"bool","bool":true},"stored":false,"revision":null}}` {
		t.Fatalf("fallback resolve: %s", got)
	}
	bad := h.Resolve(context.Background(), desktop.ResolveRequest{Scope: global(), Key: "editor.wrap", Fallback: intValue("nope")})
	requireRefusal(t, bad.Failure, "invalid", "desktop.invalid_request", desktop.CommitNone, "fallback.int", "invalid", "nope")
}

func TestW10Refusals(t *testing.T) {
	h := handler(t, memory.New())
	create(t, h, "editor.theme", "dark")
	conflict := replace(t, h, "editor.theme", text("light"), absent())
	requireRefusal(t, conflict.Failure, "conflict", "preferences.conflict", desktop.CommitNotApplied, "", "", "editor.theme", "light")
	malformed := replace(t, h, "editor.theme", intValue("abc"), revision("1"))
	requireRefusal(t, malformed.Failure, "invalid", "desktop.invalid_request", desktop.CommitNotApplied, "value.int", "invalid", "editor.theme", "abc")
	if !strings.Contains(encode(t, conflict), `"ok":false`) || conflict.Value != nil {
		t.Fatal("refusal carried a value")
	}
}

func TestW11CommitUncertainty(t *testing.T) {
	ctx := context.Background()
	failing := handler(t, storetest.FailBefore(memory.New(), storetest.Unavailable))
	outcome := replace(t, failing, "editor.theme", text("dark"), absent())
	requireRefusal(t, outcome.Failure, "unavailable", "preferences.store_unavailable", desktop.CommitUnknown, "", "")
	read := failing.Read(ctx, desktop.ReadRequest{Scope: global(), Key: "editor.theme"})
	requireRefusal(t, read.Failure, "unavailable", "preferences.store_unavailable", desktop.CommitNone, "", "")

	inner := memory.New()
	lossy := handler(t, storetest.LoseAcknowledgement(inner, storetest.Unacknowledged))
	outcome = replace(t, lossy, "editor.theme", text("dark"), absent())
	requireRefusal(t, outcome.Failure, "timeout", "preferences.store_unacknowledged", desktop.CommitUnknown, "", "")
	observed := lossy.Read(ctx, desktop.ReadRequest{Scope: global(), Key: "editor.theme"})
	if !observed.OK || !observed.Value.Found || observed.Value.Entry.Revision != "1" {
		t.Fatalf("read after lost acknowledgement: %+v", observed)
	}
	retry := replace(t, lossy, "editor.theme", text("dark"), absent())
	requireRefusal(t, retry.Failure, "conflict", "preferences.conflict", desktop.CommitNotApplied, "", "")
	reconciled := replace(t, lossy, "editor.theme", text("light"), revision(observed.Value.Entry.Revision))
	requireRefusal(t, reconciled.Failure, "timeout", "preferences.store_unacknowledged", desktop.CommitUnknown, "", "")
	final := handler(t, inner).Read(ctx, desktop.ReadRequest{Scope: global(), Key: "editor.theme"})
	if !final.OK || final.Value.Entry.Revision != "2" || *final.Value.Entry.Value.Text != "light" {
		t.Fatalf("reconciled write did not apply once: %+v", final.Value.Entry)
	}
}

func TestW12Fixtures(t *testing.T) {
	h := handler(t, memory.New())
	created := create(t, h, "editor.theme", "dark")
	if got := encode(t, created); got != createFixture {
		t.Fatalf("create fixture:\n got %s\nwant %s", got, createFixture)
	}
	conflict := replace(t, h, "editor.theme", text("light"), absent())
	if got := encode(t, conflict); got != conflictFixture {
		t.Fatalf("conflict fixture:\n got %s\nwant %s", got, conflictFixture)
	}
	var request desktop.ReplaceRequest
	if err := json.Unmarshal([]byte(`{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"int","int":5},"expected":{"kind":"absent"}}`), &request); err == nil {
		t.Fatal("a JSON number for int must not decode into the request type")
	}
}

func TestW13Documents(t *testing.T) {
	ctx := context.Background()
	h := handler(t, memory.New())
	created := h.ReplaceDocument(ctx, []byte(`{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"text","text":"dark"},"expected":{"kind":"absent"}}`))
	if got := encode(t, created); got != createFixture {
		t.Fatalf("document create:\n got %s\nwant %s", got, createFixture)
	}
	read := h.ReadDocument(ctx, []byte(`{"scope":{"kind":"global"},"key":"editor.theme"}`))
	if !read.OK || !read.Value.Found || read.Value.Entry.Revision != "1" {
		t.Fatalf("document read: %+v", read)
	}
	number := h.ReplaceDocument(ctx, []byte(`{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"int","int":5},"expected":{"kind":"absent"}}`))
	requireRefusal(t, number.Failure, "invalid", "desktop.invalid_request", desktop.CommitNotApplied, "request", "invalid")
	for _, document := range []string{`not json`, `[]`, `42`, `null`, `"{}"`} {
		write := h.ReplaceDocument(ctx, []byte(document))
		requireRefusal(t, write.Failure, "invalid", "desktop.invalid_request", desktop.CommitNotApplied, "request", "invalid", "not json")
		remove := h.RemoveDocument(ctx, []byte(document))
		requireRefusal(t, remove.Failure, "invalid", "desktop.invalid_request", desktop.CommitNotApplied, "request", "invalid")
		readBad := h.ReadDocument(ctx, []byte(document))
		requireRefusal(t, readBad.Failure, "invalid", "desktop.invalid_request", desktop.CommitNone, "request", "invalid")
		listBad := h.ListDocument(ctx, []byte(document))
		requireRefusal(t, listBad.Failure, "invalid", "desktop.invalid_request", desktop.CommitNone, "request", "invalid")
		resolveBad := h.ResolveDocument(ctx, []byte(document))
		requireRefusal(t, resolveBad.Failure, "invalid", "desktop.invalid_request", desktop.CommitNone, "request", "invalid")
	}
}
