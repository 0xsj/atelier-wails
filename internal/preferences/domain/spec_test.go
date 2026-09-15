package domain_test

import (
	"math"
	"strings"
	"testing"

	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

const (
	firstWorkspace  = "01900000-0000-7000-8000-000000000001"
	secondWorkspace = "01900000-0000-7000-8000-000000000002"
)

func parseID(t *testing.T, text string) id.ID {
	t.Helper()
	value, err := id.Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func key(t *testing.T, text string) domain.Key {
	t.Helper()
	value, err := domain.NewKey(text)
	if err != nil {
		t.Fatalf("key %q: %v", text, err)
	}
	return value
}

func text(t *testing.T, value string) domain.Value {
	t.Helper()
	result, err := domain.Text(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func revision(t *testing.T, value uint64) domain.Expected {
	t.Helper()
	result, err := domain.ExpectRevision(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func entry(t *testing.T, scope domain.Scope, k domain.Key, value domain.Value, rev uint64) domain.Entry {
	t.Helper()
	result, err := domain.Restore(scope, k, value, rev)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// requireFailure checks P15 for every refusal: shared kind, diagnostic type and
// no echo of the supplied texts.
func requireFailure(t *testing.T, err error, kind faults.Kind, typ string, secrets ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s failure", typ)
	}
	if !faults.IsKind(err, kind) {
		t.Fatalf("kind: got %v, want %v (%v)", err, kind, err)
	}
	if got := faults.DiagnosticTypeOf(err); got != typ {
		t.Fatalf("type: got %q, want %q", got, typ)
	}
	message := err.Error()
	fields := faults.FieldsOf(err)
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		if strings.Contains(message, secret) {
			t.Fatalf("message echoes input: %q", message)
		}
		for _, value := range fields {
			if strings.Contains(value, secret) {
				t.Fatalf("field echoes input: %q", value)
			}
		}
	}
}

func TestP01Scopes(t *testing.T) {
	global := domain.Global()
	if !global.Valid() || global.Kind() != domain.GlobalScope || global.String() != "global" {
		t.Fatalf("global scope: %v", global)
	}
	if _, ok := global.WorkspaceID(); ok {
		t.Fatal("global scope reports a workspace")
	}
	first, err := domain.ForWorkspace(parseID(t, firstWorkspace))
	if err != nil {
		t.Fatal(err)
	}
	if !first.Valid() || first.Kind() != domain.WorkspaceScope || first.String() != "workspace:"+firstWorkspace {
		t.Fatalf("workspace scope: %v", first)
	}
	if got, ok := first.WorkspaceID(); !ok || got != parseID(t, firstWorkspace) {
		t.Fatal("workspace ID accessor")
	}
	again, _ := domain.ForWorkspace(parseID(t, firstWorkspace))
	second, _ := domain.ForWorkspace(parseID(t, secondWorkspace))
	if first != again || first == second || first == global {
		t.Fatal("scope equality")
	}
	_, err = domain.ForWorkspace(id.ID{})
	requireFailure(t, err, faults.Invalid, "preferences.invalid_scope")
	var zero domain.Scope
	if zero.Valid() || zero.String() != "invalid" {
		t.Fatal("zero scope must be invalid")
	}
}

func TestP02Keys(t *testing.T) {
	longest := "a" + strings.Repeat("z", 127)
	for _, accepted := range []string{"a", "editor.theme", "a-b_c.d9", longest} {
		k, err := domain.NewKey(accepted)
		if err != nil || k.String() != accepted || !k.Valid() {
			t.Fatalf("accepted key %q: %v", accepted, err)
		}
	}
	refused := []string{"", longest + "z", "A", "1a", " a", "a ", "a/b", "é", "Editor.Theme"}
	for _, input := range refused {
		_, err := domain.NewKey(input)
		requireFailure(t, err, faults.Invalid, "preferences.invalid_key", input)
	}
	var zero domain.Key
	if zero.Valid() {
		t.Fatal("zero key must be invalid")
	}
}

func TestP03Values(t *testing.T) {
	empty := text(t, "")
	if got, ok := empty.Text(); !ok || got != "" || empty.Kind() != domain.TextValue {
		t.Fatal("empty text")
	}
	limit := strings.Repeat("x", 4096)
	if _, err := domain.Text(limit); err != nil {
		t.Fatal("4096-byte text refused")
	}
	_, err := domain.Text(limit + "x")
	requireFailure(t, err, faults.Invalid, "preferences.invalid_value", "xxxx")
	_, err = domain.Text("ok\xff")
	requireFailure(t, err, faults.Invalid, "preferences.invalid_value", "ok")

	flag := domain.Bool(true)
	if got, ok := flag.Bool(); !ok || !got || flag.Kind() != domain.BoolValue {
		t.Fatal("bool value")
	}
	for _, n := range []int64{math.MinInt64, math.MaxInt64} {
		v := domain.Int(n)
		if got, ok := v.Int(); !ok || got != n || v.Kind() != domain.IntValue {
			t.Fatal("int value")
		}
	}
	if _, ok := flag.Text(); ok {
		t.Fatal("text accessor on bool")
	}
	if _, ok := flag.Int(); ok {
		t.Fatal("int accessor on bool")
	}
	if _, ok := domain.Int(1).Bool(); ok {
		t.Fatal("bool accessor on int")
	}
	one := text(t, "1")
	if one.Equal(domain.Int(1)) || domain.Int(1).Equal(domain.Bool(true)) || one.Equal(domain.Bool(true)) {
		t.Fatal("cross-kind equality")
	}
	if !one.Equal(text(t, "1")) || !domain.Int(1).Equal(domain.Int(1)) || one.Equal(text(t, "2")) {
		t.Fatal("same-kind equality")
	}
	var zero domain.Value
	if zero.Valid() {
		t.Fatal("zero value must be invalid")
	}
}

func TestP04Expected(t *testing.T) {
	absent := domain.ExpectAbsent()
	if !absent.IsAbsent() || !absent.Valid() {
		t.Fatal("absent")
	}
	if _, ok := absent.Revision(); ok {
		t.Fatal("absent reports a revision")
	}
	one := revision(t, 1)
	if one.IsAbsent() {
		t.Fatal("revision reports absent")
	}
	if got, ok := one.Revision(); !ok || got != 1 {
		t.Fatal("revision accessor")
	}
	_, err := domain.ExpectRevision(0)
	requireFailure(t, err, faults.Invalid, "preferences.invalid_revision")
	var zero domain.Expected
	if !zero.IsAbsent() || !zero.Valid() {
		t.Fatal("zero expected is absent")
	}
}

func TestP05RestoreEntry(t *testing.T) {
	scope := domain.Global()
	k := key(t, "editor.theme")
	value := text(t, "dark")
	restored := entry(t, scope, k, value, 5)
	if restored.Scope() != scope || restored.Key() != k || !restored.Value().Equal(value) || restored.Revision() != 5 || !restored.Valid() {
		t.Fatal("restored entry accessors")
	}
	_, err := domain.Restore(scope, k, value, 0)
	requireFailure(t, err, faults.Invalid, "preferences.invalid_revision")
	_, err = domain.Restore(domain.Scope{}, k, value, 1)
	requireFailure(t, err, faults.Invalid, "preferences.invalid_scope")
	_, err = domain.Restore(scope, domain.Key{}, value, 1)
	requireFailure(t, err, faults.Invalid, "preferences.invalid_key")
	_, err = domain.Restore(scope, k, domain.Value{}, 1)
	requireFailure(t, err, faults.Invalid, "preferences.invalid_value")
	var zero domain.Entry
	if zero.Valid() {
		t.Fatal("zero entry must be invalid")
	}
}

func TestP06Create(t *testing.T) {
	scope := domain.Global()
	k := key(t, "editor.theme")
	value := text(t, "dark")
	result, err := domain.DecideReplace(nil, scope, k, value, domain.ExpectAbsent())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.Changed || result.Entry.Revision() != 1 || !result.Entry.Value().Equal(value) {
		t.Fatalf("create result: %+v", result)
	}
	if result.Entry.Scope() != scope || result.Entry.Key() != k {
		t.Fatal("create entry identity")
	}
	if result.Event == nil || result.Event.Name() != domain.ChangedEvent || result.Event.Revision() != 1 {
		t.Fatalf("create event: %+v", result.Event)
	}
	if got, ok := result.Event.Value(); !ok || !got.Equal(value) || result.Event.Scope() != scope || result.Event.Key() != k {
		t.Fatal("create event payload")
	}
	refused, err := domain.DecideReplace(nil, scope, k, value, revision(t, 1))
	requireFailure(t, err, faults.Conflict, "preferences.conflict", "editor.theme", "dark")
	if refused.Event != nil || refused.Status != 0 {
		t.Fatal("refusal returned a result")
	}
}

func TestP07ReplaceConflicts(t *testing.T) {
	scope := domain.Global()
	k := key(t, "editor.theme")
	current := entry(t, scope, k, text(t, "dark"), 3)
	for _, expected := range []domain.Expected{domain.ExpectAbsent(), revision(t, 2), revision(t, 4)} {
		result, err := domain.DecideReplace(&current, scope, k, text(t, "light"), expected)
		requireFailure(t, err, faults.Conflict, "preferences.conflict", "editor.theme", "light")
		if result.Event != nil || result.Status != 0 || result.Entry.Valid() {
			t.Fatal("conflict returned a result")
		}
	}
	if current.Revision() != 3 {
		t.Fatal("current mutated")
	}
}

func TestP08ReplaceUnchanged(t *testing.T) {
	scope := domain.Global()
	k := key(t, "editor.theme")
	current := entry(t, scope, k, text(t, "dark"), 3)
	result, err := domain.DecideReplace(&current, scope, k, text(t, "dark"), revision(t, 3))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.Unchanged || result.Event != nil || result.Entry != current || result.Entry.Revision() != 3 {
		t.Fatalf("unchanged: %+v", result)
	}
}

func TestP09ReplaceChanged(t *testing.T) {
	scope, _ := domain.ForWorkspace(parseID(t, firstWorkspace))
	k := key(t, "editor.theme")
	current := entry(t, scope, k, text(t, "dark"), 3)
	first, err := domain.DecideReplace(&current, scope, k, text(t, "light"), revision(t, 3))
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != domain.Changed || first.Entry.Revision() != 4 || !first.Entry.Value().Equal(text(t, "light")) {
		t.Fatalf("changed: %+v", first)
	}
	if first.Event == nil || first.Event.Name() != domain.ChangedEvent || first.Event.Revision() != 4 {
		t.Fatal("changed event")
	}
	if got, ok := first.Event.Value(); !ok || !got.Equal(text(t, "light")) {
		t.Fatal("changed event value")
	}
	if current.Revision() != 3 || !current.Value().Equal(text(t, "dark")) {
		t.Fatal("caller's current entry mutated")
	}
	next := first.Entry
	second, err := domain.DecideReplace(&next, scope, k, domain.Int(1), revision(t, 4))
	if err != nil {
		t.Fatal(err)
	}
	if second.Entry.Revision() != 5 || second.Entry.Value().Kind() != domain.IntValue {
		t.Fatal("kind change")
	}
}

func TestP10RevisionExhausted(t *testing.T) {
	scope := domain.Global()
	k := key(t, "editor.theme")
	current := entry(t, scope, k, text(t, "dark"), math.MaxUint64)
	last := revision(t, math.MaxUint64)
	_, err := domain.DecideReplace(&current, scope, k, text(t, "light"), last)
	requireFailure(t, err, faults.Conflict, "preferences.revision_exhausted", "editor.theme", "light")
	same, err := domain.DecideReplace(&current, scope, k, text(t, "dark"), last)
	if err != nil || same.Status != domain.Unchanged || same.Event != nil {
		t.Fatal("equal replace at maximum revision")
	}
	_, err = domain.DecideRemove(&current, scope, k, last)
	requireFailure(t, err, faults.Conflict, "preferences.revision_exhausted", "editor.theme")
	if current.Revision() != math.MaxUint64 {
		t.Fatal("current mutated")
	}
}

func TestP11RemoveAbsent(t *testing.T) {
	scope := domain.Global()
	k := key(t, "editor.theme")
	for _, expected := range []domain.Expected{domain.ExpectAbsent(), revision(t, 7)} {
		result, err := domain.DecideRemove(nil, scope, k, expected)
		if err != nil {
			t.Fatal(err)
		}
		if result.Removed || result.Revision != 0 || result.Event != nil {
			t.Fatalf("absent remove: %+v", result)
		}
	}
}

func TestP12RemovePresent(t *testing.T) {
	scope := domain.Global()
	k := key(t, "editor.theme")
	current := entry(t, scope, k, text(t, "dark"), 3)
	for _, expected := range []domain.Expected{domain.ExpectAbsent(), revision(t, 2)} {
		result, err := domain.DecideRemove(&current, scope, k, expected)
		requireFailure(t, err, faults.Conflict, "preferences.conflict", "editor.theme", "dark")
		if result.Removed || result.Event != nil {
			t.Fatal("conflict returned a result")
		}
	}
	result, err := domain.DecideRemove(&current, scope, k, revision(t, 3))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Removed || result.Revision != 4 || result.Event == nil {
		t.Fatalf("removed: %+v", result)
	}
	if result.Event.Name() != domain.RemovedEvent || result.Event.Revision() != 4 || result.Event.Scope() != scope || result.Event.Key() != k {
		t.Fatal("removed event")
	}
	if _, ok := result.Event.Value(); ok {
		t.Fatal("removed event carries a value")
	}
	if current.Revision() != 3 {
		t.Fatal("current mutated")
	}
}

func TestP13MismatchedCurrent(t *testing.T) {
	scope := domain.Global()
	other, _ := domain.ForWorkspace(parseID(t, firstWorkspace))
	k := key(t, "editor.theme")
	otherKey := key(t, "editor.font")
	value := text(t, "dark")
	wrongKey := entry(t, scope, otherKey, value, 1)
	wrongScope := entry(t, other, k, value, 1)
	var invalidCurrent domain.Entry
	for _, current := range []domain.Entry{wrongKey, wrongScope, invalidCurrent} {
		current := current
		_, err := domain.DecideReplace(&current, scope, k, value, revision(t, 1))
		requireFailure(t, err, faults.Invalid, "preferences.invalid_entry", "editor", "dark")
		_, err = domain.DecideRemove(&current, scope, k, revision(t, 1))
		requireFailure(t, err, faults.Invalid, "preferences.invalid_entry", "editor", "dark")
	}
}

func TestP14ValidationPrecedesState(t *testing.T) {
	scope := domain.Global()
	k := key(t, "editor.theme")
	current := entry(t, scope, k, text(t, "dark"), 3)
	var badKey domain.Key
	// A valid key would conflict here (expected absent against revision 3).
	_, err := domain.DecideReplace(&current, scope, badKey, text(t, "light"), domain.ExpectAbsent())
	requireFailure(t, err, faults.Invalid, "preferences.invalid_key", "light")
	_, err = domain.DecideRemove(&current, scope, badKey, domain.ExpectAbsent())
	requireFailure(t, err, faults.Invalid, "preferences.invalid_key")
	_, err = domain.DecideReplace(&current, domain.Scope{}, k, text(t, "light"), domain.ExpectAbsent())
	requireFailure(t, err, faults.Invalid, "preferences.invalid_scope", "light")
	_, err = domain.DecideReplace(&current, scope, k, domain.Value{}, domain.ExpectAbsent())
	requireFailure(t, err, faults.Invalid, "preferences.invalid_value")
	if current.Revision() != 3 {
		t.Fatal("current mutated")
	}
}
