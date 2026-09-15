// Package desktop is the plain-data wire boundary for Preferences: JSON
// request and response shapes, request decoding, outcome encoding and public
// failure projection with commit status. It imports no native framework; the
// host facade forwards to Handler. See CONTRACT.md.
package desktop

// Kind names used by tagged wire shapes.
const (
	ScopeGlobal    = "global"
	ScopeWorkspace = "workspace"
	ValueText      = "text"
	ValueBool      = "bool"
	ValueInt       = "int"
	ExpectedAbsent = "absent"
	ExpectedRev    = "revision"
	StatusChanged  = "changed"
	StatusSame     = "unchanged"
)

// Commit status values carried by every failure.
const (
	CommitNone       = "none"
	CommitNotApplied = "not_applied"
	CommitUnknown    = "unknown"
)

// Scope is {"kind":"global"} or {"kind":"workspace","workspace_id":"<uuid>"}.
type Scope struct {
	Kind        string `json:"kind"`
	WorkspaceID string `json:"workspace_id,omitempty"`
}

// Value is one of the three tagged value shapes; 64-bit ints are decimal strings.
type Value struct {
	Kind string  `json:"kind"`
	Text *string `json:"text,omitempty"`
	Bool *bool   `json:"bool,omitempty"`
	Int  *string `json:"int,omitempty"`
}

// Expected is {"kind":"absent"} or {"kind":"revision","revision":"3"}.
type Expected struct {
	Kind     string  `json:"kind"`
	Revision *string `json:"revision,omitempty"`
}

// Entry is a stored preference; Revision is a decimal string.
type Entry struct {
	Scope    Scope  `json:"scope"`
	Key      string `json:"key"`
	Value    Value  `json:"value"`
	Revision string `json:"revision"`
}

// Event is a returned fact; Value is null for a removal.
type Event struct {
	Name     string `json:"name"`
	Scope    Scope  `json:"scope"`
	Key      string `json:"key"`
	Revision string `json:"revision"`
	Value    *Value `json:"value"`
}

// Failure is the public projection plus commit status.
type Failure struct {
	Kind    string            `json:"kind"`
	Message string            `json:"message"`
	Type    *string           `json:"type"`
	Fields  map[string]string `json:"fields"`
	Commit  string            `json:"commit"`
}

type ReadRequest struct {
	Scope Scope  `json:"scope"`
	Key   string `json:"key"`
}

type ListRequest struct {
	Scope Scope `json:"scope"`
}

type ReplaceRequest struct {
	Scope    Scope    `json:"scope"`
	Key      string   `json:"key"`
	Value    Value    `json:"value"`
	Expected Expected `json:"expected"`
}

type RemoveRequest struct {
	Scope    Scope    `json:"scope"`
	Key      string   `json:"key"`
	Expected Expected `json:"expected"`
}

type ResolveRequest struct {
	Scope    Scope  `json:"scope"`
	Key      string `json:"key"`
	Fallback Value  `json:"fallback"`
}

type ReadResponse struct {
	Found bool   `json:"found"`
	Entry *Entry `json:"entry"`
}

type ListResponse struct {
	Entries []Entry `json:"entries"`
}

type ReplaceResponse struct {
	Status string `json:"status"`
	Entry  Entry  `json:"entry"`
	Event  *Event `json:"event"`
}

type RemoveResponse struct {
	Removed  bool    `json:"removed"`
	Revision *string `json:"revision"`
	Event    *Event  `json:"event"`
}

type ResolveResponse struct {
	Value    Value   `json:"value"`
	Stored   bool    `json:"stored"`
	Revision *string `json:"revision"`
}

// Outcomes carry exactly one of Value or Failure. They are separate types so
// the native binding layer sees concrete, non-generic shapes.

type ReadOutcome struct {
	OK      bool          `json:"ok"`
	Value   *ReadResponse `json:"value,omitempty"`
	Failure *Failure      `json:"failure,omitempty"`
}

type ListOutcome struct {
	OK      bool          `json:"ok"`
	Value   *ListResponse `json:"value,omitempty"`
	Failure *Failure      `json:"failure,omitempty"`
}

type ReplaceOutcome struct {
	OK      bool             `json:"ok"`
	Value   *ReplaceResponse `json:"value,omitempty"`
	Failure *Failure         `json:"failure,omitempty"`
}

type RemoveOutcome struct {
	OK      bool            `json:"ok"`
	Value   *RemoveResponse `json:"value,omitempty"`
	Failure *Failure        `json:"failure,omitempty"`
}

type ResolveOutcome struct {
	OK      bool             `json:"ok"`
	Value   *ResolveResponse `json:"value,omitempty"`
	Failure *Failure         `json:"failure,omitempty"`
}
