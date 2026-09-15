// Package domain contains the pure Preferences vocabulary and its
// compare-and-replace decisions. It performs no storage, clock, ID generation,
// logging or host call: a caller supplies the current entry for one scope and
// key, and the domain returns decided values and returned event facts.
//
// Zero values of Scope, Key, Value and Entry are constructible in Go and are
// invalid; every decision validates its inputs before inspecting state. See
// CONTRACT.md for the promised scenarios.
package domain

import (
	"math"
	"unicode/utf8"

	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

// Event names returned by successful transitions.
const (
	ChangedEvent = "preference.changed"
	RemovedEvent = "preference.removed"
)

const (
	maxKeyLength   = 128
	maxTextBytes   = 4096
	maxRevision    = math.MaxUint64
	invalidDisplay = "invalid"
)

// ScopeKind distinguishes the global installation scope from a workspace scope.
type ScopeKind uint8

const (
	GlobalScope ScopeKind = iota + 1
	WorkspaceScope
)

// Scope is either global or one opaque workspace ID. The zero Scope is invalid.
type Scope struct {
	kind      ScopeKind
	workspace id.ID
}

func Global() Scope { return Scope{kind: GlobalScope} }

// ForWorkspace refuses the zero ID; it does not check that the workspace exists.
func ForWorkspace(workspace id.ID) (Scope, error) {
	if workspace.IsZero() {
		return Scope{}, invalid("preferences.invalid_scope")
	}
	return Scope{kind: WorkspaceScope, workspace: workspace}, nil
}

func (s Scope) Kind() ScopeKind { return s.kind }

// WorkspaceID returns the workspace ID only for a valid workspace scope.
func (s Scope) WorkspaceID() (id.ID, bool) {
	if s.kind != WorkspaceScope || s.workspace.IsZero() {
		return id.ID{}, false
	}
	return s.workspace, true
}

func (s Scope) Valid() bool {
	switch s.kind {
	case GlobalScope:
		return s.workspace.IsZero()
	case WorkspaceScope:
		return !s.workspace.IsZero()
	default:
		return false
	}
}

func (s Scope) String() string {
	if !s.Valid() {
		return invalidDisplay
	}
	if s.kind == GlobalScope {
		return "global"
	}
	return "workspace:" + s.workspace.String()
}

// Key is validated lowercase ASCII matching [a-z][a-z0-9_.-]{0,127}. Its field
// is private so a conversion cannot bypass NewKey; the zero Key is invalid.
type Key struct{ value string }

func NewKey(value string) (Key, error) {
	if !validKey(value) {
		return Key{}, invalid("preferences.invalid_key")
	}
	return Key{value: value}, nil
}

func (k Key) String() string { return k.value }
func (k Key) Valid() bool    { return validKey(k.value) }

func validKey(value string) bool {
	if len(value) == 0 || len(value) > maxKeyLength {
		return false
	}
	if value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for i := 1; i < len(value); i++ {
		b := value[i]
		lower := b >= 'a' && b <= 'z'
		digit := b >= '0' && b <= '9'
		if !lower && !digit && b != '.' && b != '_' && b != '-' {
			return false
		}
	}
	return true
}

// ValueKind is the type of a preference value; it is fixed by the value supplied.
type ValueKind uint8

const (
	TextValue ValueKind = iota + 1
	BoolValue
	IntValue
)

// Value is UTF-8 text of at most 4096 bytes, a boolean or a signed 64-bit
// integer. The zero Value is invalid.
type Value struct {
	kind ValueKind
	text string
	flag bool
	intv int64
}

func Text(value string) (Value, error) {
	if !validText(value) {
		return Value{}, invalid("preferences.invalid_value")
	}
	return Value{kind: TextValue, text: value}, nil
}

func Bool(value bool) Value { return Value{kind: BoolValue, flag: value} }
func Int(value int64) Value { return Value{kind: IntValue, intv: value} }

func (v Value) Kind() ValueKind { return v.kind }

// Text returns the payload only for a text value.
func (v Value) Text() (string, bool) {
	if v.kind != TextValue {
		return "", false
	}
	return v.text, true
}

// Bool returns the payload only for a boolean value.
func (v Value) Bool() (bool, bool) {
	if v.kind != BoolValue {
		return false, false
	}
	return v.flag, true
}

// Int returns the payload only for an integer value.
func (v Value) Int() (int64, bool) {
	if v.kind != IntValue {
		return 0, false
	}
	return v.intv, true
}

func (v Value) Valid() bool {
	switch v.kind {
	case TextValue:
		return validText(v.text)
	case BoolValue, IntValue:
		return true
	default:
		return false
	}
}

// Equal compares kind and payload; values of different kinds are never equal.
func (v Value) Equal(other Value) bool {
	if v.kind != other.kind {
		return false
	}
	switch v.kind {
	case TextValue:
		return v.text == other.text
	case BoolValue:
		return v.flag == other.flag
	case IntValue:
		return v.intv == other.intv
	default:
		return false
	}
}

func validText(value string) bool {
	return len(value) <= maxTextBytes && utf8.ValidString(value)
}

// Expected is absent-for-create or an exact existing revision. The zero
// Expected means absent.
type Expected struct {
	exists   bool
	revision uint64
}

func ExpectAbsent() Expected { return Expected{} }

func ExpectRevision(revision uint64) (Expected, error) {
	if revision == 0 {
		return Expected{}, invalid("preferences.invalid_revision")
	}
	return Expected{exists: true, revision: revision}, nil
}

func (e Expected) IsAbsent() bool { return !e.exists }

// Revision returns the expected revision only when an existing entry is expected.
func (e Expected) Revision() (uint64, bool) {
	if !e.exists {
		return 0, false
	}
	return e.revision, true
}

func (e Expected) Valid() bool { return !e.exists || e.revision != 0 }

func (e Expected) matches(revision uint64) bool {
	return e.exists && e.revision == revision
}

// Entry is a stored preference: scope, key, value and a positive revision.
// Fields are private; adapters rebuild entries through Restore.
type Entry struct {
	scope    Scope
	key      Key
	value    Value
	revision uint64
}

// Restore performs validated construction for adapters decoding stored records.
func Restore(scope Scope, key Key, value Value, revision uint64) (Entry, error) {
	if err := validateInput(scope, key, value); err != nil {
		return Entry{}, err
	}
	if revision == 0 {
		return Entry{}, invalid("preferences.invalid_revision")
	}
	return Entry{scope: scope, key: key, value: value, revision: revision}, nil
}

func (e Entry) Scope() Scope     { return e.scope }
func (e Entry) Key() Key         { return e.key }
func (e Entry) Value() Value     { return e.value }
func (e Entry) Revision() uint64 { return e.revision }

func (e Entry) Valid() bool {
	return e.scope.Valid() && e.key.Valid() && e.value.Valid() && e.revision != 0
}

// Event is a returned fact describing a committed-to-be transition. The
// application publishes it after its store transaction commits.
type Event struct {
	name     string
	scope    Scope
	key      Key
	value    Value
	hasValue bool
	revision uint64
}

func (e Event) Name() string     { return e.name }
func (e Event) Scope() Scope     { return e.scope }
func (e Event) Key() Key         { return e.key }
func (e Event) Revision() uint64 { return e.revision }

// Value returns the new value only for a changed event.
func (e Event) Value() (Value, bool) {
	if !e.hasValue {
		return Value{}, false
	}
	return e.value, true
}

// ChangeStatus reports whether a replace decision changed the entry.
type ChangeStatus uint8

const (
	Changed ChangeStatus = iota + 1
	Unchanged
)

// ReplaceResult is Changed with an event, or Unchanged with the current entry
// and a nil Event.
type ReplaceResult struct {
	Status ChangeStatus
	Entry  Entry
	Event  *Event
}

// RemoveResult is Removed with the removal revision and event, or the absent
// no-op with Removed false and a nil Event.
type RemoveResult struct {
	Removed  bool
	Revision uint64
	Event    *Event
}

// DecideReplace validates inputs, then compares expected against current and
// returns the entry and event a store should commit. It never mutates current.
func DecideReplace(current *Entry, scope Scope, key Key, value Value, expected Expected) (ReplaceResult, error) {
	if err := validateInput(scope, key, value); err != nil {
		return ReplaceResult{}, err
	}
	if !expected.Valid() {
		return ReplaceResult{}, invalid("preferences.invalid_revision")
	}
	if current == nil {
		if !expected.IsAbsent() {
			return ReplaceResult{}, conflict()
		}
		entry := Entry{scope: scope, key: key, value: value, revision: 1}
		return ReplaceResult{Status: Changed, Entry: entry, Event: changedEvent(entry)}, nil
	}
	if err := validateCurrent(*current, scope, key); err != nil {
		return ReplaceResult{}, err
	}
	if !expected.matches(current.revision) {
		return ReplaceResult{}, conflict()
	}
	if current.value.Equal(value) {
		return ReplaceResult{Status: Unchanged, Entry: *current}, nil
	}
	revision, err := nextRevision(current.revision)
	if err != nil {
		return ReplaceResult{}, err
	}
	entry := Entry{scope: scope, key: key, value: value, revision: revision}
	return ReplaceResult{Status: Changed, Entry: entry, Event: changedEvent(entry)}, nil
}

// DecideRemove validates inputs, treats an absent entry as a successful no-op,
// and otherwise requires the exact current revision. It never mutates current.
func DecideRemove(current *Entry, scope Scope, key Key, expected Expected) (RemoveResult, error) {
	if !scope.Valid() {
		return RemoveResult{}, invalid("preferences.invalid_scope")
	}
	if !key.Valid() {
		return RemoveResult{}, invalid("preferences.invalid_key")
	}
	if !expected.Valid() {
		return RemoveResult{}, invalid("preferences.invalid_revision")
	}
	if current == nil {
		return RemoveResult{}, nil
	}
	if err := validateCurrent(*current, scope, key); err != nil {
		return RemoveResult{}, err
	}
	if !expected.matches(current.revision) {
		return RemoveResult{}, conflict()
	}
	revision, err := nextRevision(current.revision)
	if err != nil {
		return RemoveResult{}, err
	}
	event := &Event{name: RemovedEvent, scope: scope, key: key, revision: revision}
	return RemoveResult{Removed: true, Revision: revision, Event: event}, nil
}

func changedEvent(entry Entry) *Event {
	return &Event{
		name:     ChangedEvent,
		scope:    entry.scope,
		key:      entry.key,
		value:    entry.value,
		hasValue: true,
		revision: entry.revision,
	}
}

func validateInput(scope Scope, key Key, value Value) error {
	if !scope.Valid() {
		return invalid("preferences.invalid_scope")
	}
	if !key.Valid() {
		return invalid("preferences.invalid_key")
	}
	if !value.Valid() {
		return invalid("preferences.invalid_value")
	}
	return nil
}

func validateCurrent(current Entry, scope Scope, key Key) error {
	if !current.Valid() || current.scope != scope || current.key != key {
		return invalid("preferences.invalid_entry")
	}
	return nil
}

func nextRevision(current uint64) (uint64, error) {
	if current == maxRevision {
		return 0, faults.New(faults.Conflict, "preference revision exhausted").
			WithType("preferences.revision_exhausted")
	}
	return current + 1, nil
}

func invalid(typ string) error {
	return faults.New(faults.Invalid, "invalid preference input").WithType(typ)
}

func conflict() error {
	return faults.New(faults.Conflict, "preference revision conflict").WithType("preferences.conflict")
}
