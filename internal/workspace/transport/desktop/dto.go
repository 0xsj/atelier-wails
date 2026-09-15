// Package desktop is the plain-data wire boundary for Workspace metadata and
// lifecycle operations. It imports no Wails type; the host facade forwards to
// Handler. See CONTRACT.md.
package desktop

const (
	FilterActive     = "active"
	FilterAll        = "all"
	ExpectedAbsent   = "absent"
	ExpectedRevision = "revision"
	StatusChanged    = "changed"
	StatusUnchanged  = "unchanged"
)

const (
	CommitNone       = "none"
	CommitNotApplied = "not_applied"
	CommitUnknown    = "unknown"
)

type Expected struct {
	Kind     string  `json:"kind"`
	Revision *string `json:"revision,omitempty"`
}

type ReadRequest struct {
	ID string `json:"id"`
}
type ListRequest struct {
	Filter string `json:"filter"`
}

type RegisterRequest struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	At       string `json:"at"`
}

type RenameRequest struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Expected Expected `json:"expected"`
	At       string   `json:"at"`
}

type LifecycleRequest struct {
	ID       string   `json:"id"`
	Expected Expected `json:"expected"`
	At       *string  `json:"at,omitempty"`
}

type Workspace struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Location  string `json:"location"`
	Status    string `json:"status"`
	Revision  string `json:"revision"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Event struct {
	Name        string `json:"name"`
	WorkspaceID string `json:"workspace_id"`
	Revision    string `json:"revision"`
	Title       string `json:"title"`
	Location    string `json:"location"`
	Status      string `json:"status"`
}

type ReadResponse struct {
	Found     bool       `json:"found"`
	Workspace *Workspace `json:"workspace,omitempty"`
}
type ListResponse struct {
	Workspaces []Workspace `json:"workspaces"`
}
type RegisteredResponse struct {
	Workspace Workspace `json:"workspace"`
	Event     Event     `json:"event"`
}
type MutationResponse struct {
	Status    string    `json:"status"`
	Workspace Workspace `json:"workspace"`
	Event     *Event    `json:"event"`
}
type ForgottenResponse struct {
	Revision string `json:"revision"`
	Event    Event  `json:"event"`
}
type ForgetResponse struct {
	Found     bool               `json:"found"`
	Workspace *ForgottenResponse `json:"workspace,omitempty"`
}

type Failure struct {
	Kind    string            `json:"kind"`
	Message string            `json:"message"`
	Type    *string           `json:"type"`
	Fields  map[string]string `json:"fields"`
	Commit  string            `json:"commit"`
}

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
type RegisteredOutcome struct {
	OK      bool                `json:"ok"`
	Value   *RegisteredResponse `json:"value,omitempty"`
	Failure *Failure            `json:"failure,omitempty"`
}
type MutationOutcome struct {
	OK      bool              `json:"ok"`
	Value   *MutationResponse `json:"value,omitempty"`
	Failure *Failure          `json:"failure,omitempty"`
}
type ForgetOutcome struct {
	OK      bool            `json:"ok"`
	Value   *ForgetResponse `json:"value,omitempty"`
	Failure *Failure        `json:"failure,omitempty"`
}
