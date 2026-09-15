// Package domain contains the pure Workspace vocabulary and lifecycle
// decisions: register, rename, archive, restore and forget. It reads no
// clock or filesystem and generates no ID; callers supply the ID, the
// timestamp and a canonical location string. See CONTRACT.md.
package domain

import (
	"math"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

// Event names returned by successful transitions.
const (
	RegisteredEvent = "workspace.registered"
	RenamedEvent    = "workspace.renamed"
	ArchivedEvent   = "workspace.archived"
	RestoredEvent   = "workspace.restored"
	ForgottenEvent  = "workspace.forgotten"
)

const maxNameLength = 120

// Status is the lifecycle state; the zero Status is invalid.
type Status uint8

const (
	Active Status = iota + 1
	Archived
)

func (s Status) Valid() bool { return s == Active || s == Archived }

func (s Status) String() string {
	switch s {
	case Active:
		return "active"
	case Archived:
		return "archived"
	default:
		return "invalid"
	}
}

// ListFilter selects which workspaces a list should return.
type ListFilter uint8

const (
	All ListFilter = iota + 1
	ActiveOnly
)

func (f ListFilter) Valid() bool { return f == All || f == ActiveOnly }

// Expected is absent-for-create or an exact existing revision. The zero
// Expected means absent.
type Expected struct {
	exists   bool
	revision uint64
}

func ExpectAbsent() Expected { return Expected{} }

func ExpectRevision(revision uint64) (Expected, error) {
	if revision == 0 {
		return Expected{}, invalid("workspace.invalid_revision")
	}
	return Expected{exists: true, revision: revision}, nil
}

func (e Expected) IsAbsent() bool { return !e.exists }

func (e Expected) Revision() (uint64, bool) {
	if !e.exists {
		return 0, false
	}
	return e.revision, true
}

func (e Expected) Valid() bool { return !e.exists || e.revision != 0 }

func (e Expected) matches(revision uint64) bool { return e.exists && e.revision == revision }

// Workspace is registry metadata. Fields are private; adapters rebuild a
// workspace through Rebuild. The zero Workspace is invalid.
type Workspace struct {
	id        id.ID
	name      string
	location  string
	status    Status
	revision  uint64
	createdAt time.Time
	updatedAt time.Time
}

// Rebuild performs validated construction for adapters decoding stored
// records. Timestamps are normalized to UTC milliseconds.
func Rebuild(workspaceID id.ID, name, location string, status Status, revision uint64, createdAt, updatedAt time.Time) (Workspace, error) {
	if err := validateIdentity(workspaceID, name, location); err != nil {
		return Workspace{}, err
	}
	if !status.Valid() {
		return Workspace{}, invalid("workspace.invalid_status")
	}
	if revision == 0 {
		return Workspace{}, invalid("workspace.invalid_revision")
	}
	return Workspace{
		id:        workspaceID,
		name:      name,
		location:  location,
		status:    status,
		revision:  revision,
		createdAt: normalize(createdAt),
		updatedAt: normalize(updatedAt),
	}, nil
}

func (w Workspace) ID() id.ID            { return w.id }
func (w Workspace) Name() string         { return w.name }
func (w Workspace) Location() string     { return w.location }
func (w Workspace) Status() Status       { return w.status }
func (w Workspace) Revision() uint64     { return w.revision }
func (w Workspace) CreatedAt() time.Time { return w.createdAt }
func (w Workspace) UpdatedAt() time.Time { return w.updatedAt }

func (w Workspace) Valid() bool {
	return !w.id.IsZero() && validName(w.name) && validLocation(w.location) && w.status.Valid() && w.revision != 0
}

// Event is a returned fact describing a transition the caller should commit.
type Event struct {
	name      string
	workspace id.ID
	revision  uint64
	title     string
	location  string
	status    Status
}

func (e Event) Name() string     { return e.name }
func (e Event) Workspace() id.ID { return e.workspace }
func (e Event) Revision() uint64 { return e.revision }
func (e Event) Title() string    { return e.title }
func (e Event) Location() string { return e.location }
func (e Event) Status() Status   { return e.status }

// ChangeStatus reports whether a decision changed the workspace.
type ChangeStatus uint8

const (
	Changed ChangeStatus = iota + 1
	Unchanged
)

// MutationResult is Changed with an event, or Unchanged with the current
// workspace and a nil Event.
type MutationResult struct {
	Status    ChangeStatus
	Workspace Workspace
	Event     *Event
}

// RegisterResult is a new active workspace and its event.
type RegisterResult struct {
	Workspace Workspace
	Event     Event
}

// ForgetResult is the removal revision and its event.
type ForgetResult struct {
	Revision uint64
	Event    Event
}

// Register validates inputs and returns the active workspace at revision 1.
func Register(workspaceID id.ID, name, location string, at time.Time) (RegisterResult, error) {
	if err := validateIdentity(workspaceID, name, location); err != nil {
		return RegisterResult{}, err
	}
	instant := normalize(at)
	workspace := Workspace{
		id:        workspaceID,
		name:      name,
		location:  location,
		status:    Active,
		revision:  1,
		createdAt: instant,
		updatedAt: instant,
	}
	return RegisterResult{Workspace: workspace, Event: event(RegisteredEvent, workspace)}, nil
}

// Rename changes the display name; the same name is Unchanged.
func Rename(current Workspace, name string, expected Expected, at time.Time) (MutationResult, error) {
	if !validName(name) {
		return MutationResult{}, invalid("workspace.invalid_name")
	}
	if err := validateCurrent(current, expected); err != nil {
		return MutationResult{}, err
	}
	if current.name == name {
		return MutationResult{Status: Unchanged, Workspace: current}, nil
	}
	revision, err := nextRevision(current.revision)
	if err != nil {
		return MutationResult{}, err
	}
	next := current
	next.name, next.revision, next.updatedAt = name, revision, normalize(at)
	return changed(next, RenamedEvent), nil
}

// Archive hides a workspace from the active list; already archived is Unchanged.
func Archive(current Workspace, expected Expected, at time.Time) (MutationResult, error) {
	return transition(current, expected, at, Archived, ArchivedEvent)
}

// Restore returns an archived workspace to active; already active is Unchanged.
func Restore(current Workspace, expected Expected, at time.Time) (MutationResult, error) {
	return transition(current, expected, at, Active, RestoredEvent)
}

// Forget removes registry metadata; only an archived workspace may be forgotten.
func Forget(current Workspace, expected Expected) (ForgetResult, error) {
	if err := validateCurrent(current, expected); err != nil {
		return ForgetResult{}, err
	}
	if current.status != Archived {
		return ForgetResult{}, faults.New(faults.Conflict, "active workspace cannot be forgotten").
			WithType("workspace.active_forget")
	}
	revision, err := nextRevision(current.revision)
	if err != nil {
		return ForgetResult{}, err
	}
	forgotten := current
	forgotten.revision = revision
	return ForgetResult{Revision: revision, Event: event(ForgottenEvent, forgotten)}, nil
}

func transition(current Workspace, expected Expected, at time.Time, target Status, name string) (MutationResult, error) {
	if err := validateCurrent(current, expected); err != nil {
		return MutationResult{}, err
	}
	if current.status == target {
		return MutationResult{Status: Unchanged, Workspace: current}, nil
	}
	revision, err := nextRevision(current.revision)
	if err != nil {
		return MutationResult{}, err
	}
	next := current
	next.status, next.revision, next.updatedAt = target, revision, normalize(at)
	return changed(next, name), nil
}

func changed(next Workspace, name string) MutationResult {
	ev := event(name, next)
	return MutationResult{Status: Changed, Workspace: next, Event: &ev}
}

func event(name string, workspace Workspace) Event {
	return Event{
		name:      name,
		workspace: workspace.id,
		revision:  workspace.revision,
		title:     workspace.name,
		location:  workspace.location,
		status:    workspace.status,
	}
}

func validateIdentity(workspaceID id.ID, name, location string) error {
	if workspaceID.IsZero() {
		return invalid("workspace.invalid_id")
	}
	if !validName(name) {
		return invalid("workspace.invalid_name")
	}
	if !validLocation(location) {
		return invalid("workspace.invalid_location")
	}
	return nil
}

func validateCurrent(current Workspace, expected Expected) error {
	if !current.Valid() {
		return invalid("workspace.invalid_record")
	}
	if !expected.Valid() {
		return invalid("workspace.invalid_revision")
	}
	if !expected.matches(current.revision) {
		return conflict()
	}
	return nil
}

func validName(value string) bool {
	if value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > maxNameLength {
		return false
	}
	first, _ := utf8.DecodeRuneInString(value)
	last, _ := utf8.DecodeLastRuneInString(value)
	if unicode.IsSpace(first) || unicode.IsSpace(last) {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func validLocation(value string) bool {
	if value == "" || !utf8.ValidString(value) || strings.IndexByte(value, 0) >= 0 {
		return false
	}
	return absolute(value)
}

func absolute(value string) bool {
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, `\\`) {
		return true
	}
	if len(value) < 3 {
		return false
	}
	drive := value[0]
	letter := (drive >= 'a' && drive <= 'z') || (drive >= 'A' && drive <= 'Z')
	return letter && value[1] == ':' && (value[2] == '/' || value[2] == '\\')
}

func normalize(value time.Time) time.Time { return value.UTC().Truncate(time.Millisecond) }

func nextRevision(current uint64) (uint64, error) {
	if current == math.MaxUint64 {
		return 0, faults.New(faults.Conflict, "workspace revision exhausted").
			WithType("workspace.revision_exhausted")
	}
	return current + 1, nil
}

func invalid(typ string) error {
	return faults.New(faults.Invalid, "invalid workspace input").WithType(typ)
}

func conflict() error {
	return faults.New(faults.Conflict, "workspace revision conflict").WithType("workspace.conflict")
}
