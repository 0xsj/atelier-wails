package desktop

import (
	"strconv"

	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

// Problem words a decode failure attaches to a request path.
const (
	problemMissing = "missing"
	problemUnknown = "unknown"
	problemInvalid = "invalid"
)

// invalidRequest is the structural refusal; the path never carries a value.
func invalidRequest(path, problem string) error {
	return faults.New(faults.Invalid, "invalid request").
		WithType("desktop.invalid_request").
		WithField(path, problem)
}

func decodeScope(path string, wire Scope) (domain.Scope, error) {
	switch wire.Kind {
	case "":
		return domain.Scope{}, invalidRequest(path+".kind", problemMissing)
	case ScopeGlobal:
		return domain.Global(), nil
	case ScopeWorkspace:
		if wire.WorkspaceID == "" {
			return domain.Scope{}, invalidRequest(path+".workspace_id", problemMissing)
		}
		parsed, err := id.Parse(wire.WorkspaceID)
		if err != nil {
			return domain.Scope{}, invalidRequest(path+".workspace_id", problemInvalid)
		}
		return domain.ForWorkspace(parsed)
	default:
		return domain.Scope{}, invalidRequest(path+".kind", problemUnknown)
	}
}

func decodeKey(text string) (domain.Key, error) {
	return domain.NewKey(text)
}

func decodeValue(path string, wire Value) (domain.Value, error) {
	switch wire.Kind {
	case "":
		return domain.Value{}, invalidRequest(path+".kind", problemMissing)
	case ValueText:
		if wire.Text == nil {
			return domain.Value{}, invalidRequest(path+".text", problemMissing)
		}
		return domain.Text(*wire.Text)
	case ValueBool:
		if wire.Bool == nil {
			return domain.Value{}, invalidRequest(path+".bool", problemMissing)
		}
		return domain.Bool(*wire.Bool), nil
	case ValueInt:
		if wire.Int == nil {
			return domain.Value{}, invalidRequest(path+".int", problemMissing)
		}
		n, ok := parseInt(*wire.Int)
		if !ok {
			return domain.Value{}, invalidRequest(path+".int", problemInvalid)
		}
		return domain.Int(n), nil
	default:
		return domain.Value{}, invalidRequest(path+".kind", problemUnknown)
	}
}

func decodeExpected(path string, wire Expected) (domain.Expected, error) {
	switch wire.Kind {
	case "":
		return domain.Expected{}, invalidRequest(path+".kind", problemMissing)
	case ExpectedAbsent:
		return domain.ExpectAbsent(), nil
	case ExpectedRev:
		if wire.Revision == nil {
			return domain.Expected{}, invalidRequest(path+".revision", problemMissing)
		}
		n, ok := parseRevision(*wire.Revision)
		if !ok {
			return domain.Expected{}, invalidRequest(path+".revision", problemInvalid)
		}
		return domain.ExpectRevision(n)
	default:
		return domain.Expected{}, invalidRequest(path+".kind", problemUnknown)
	}
}

// parseInt accepts -?[0-9]+ within int64; no sign prefix '+', no spaces.
func parseInt(text string) (int64, bool) {
	digits := text
	if len(digits) > 0 && digits[0] == '-' {
		digits = digits[1:]
	}
	if !allDigits(digits) {
		return 0, false
	}
	n, err := strconv.ParseInt(text, 10, 64)
	return n, err == nil
}

// parseRevision accepts [0-9]+ within uint64; zero is left to the domain.
func parseRevision(text string) (uint64, bool) {
	if !allDigits(text) {
		return 0, false
	}
	n, err := strconv.ParseUint(text, 10, 64)
	return n, err == nil
}

func allDigits(text string) bool {
	if text == "" {
		return false
	}
	for i := 0; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

func encodeScope(scope domain.Scope) Scope {
	if workspace, ok := scope.WorkspaceID(); ok {
		return Scope{Kind: ScopeWorkspace, WorkspaceID: workspace.String()}
	}
	return Scope{Kind: ScopeGlobal}
}

func encodeValue(value domain.Value) Value {
	switch value.Kind() {
	case domain.TextValue:
		text, _ := value.Text()
		return Value{Kind: ValueText, Text: &text}
	case domain.BoolValue:
		flag, _ := value.Bool()
		return Value{Kind: ValueBool, Bool: &flag}
	default:
		n, _ := value.Int()
		text := strconv.FormatInt(n, 10)
		return Value{Kind: ValueInt, Int: &text}
	}
}

func encodeRevision(revision uint64) string { return strconv.FormatUint(revision, 10) }

func encodeEntry(entry domain.Entry) Entry {
	return Entry{
		Scope:    encodeScope(entry.Scope()),
		Key:      entry.Key().String(),
		Value:    encodeValue(entry.Value()),
		Revision: encodeRevision(entry.Revision()),
	}
}

func encodeEvent(event *domain.Event) *Event {
	if event == nil {
		return nil
	}
	wire := &Event{
		Name:     event.Name(),
		Scope:    encodeScope(event.Scope()),
		Key:      event.Key().String(),
		Revision: encodeRevision(event.Revision()),
	}
	if value, ok := event.Value(); ok {
		encoded := encodeValue(value)
		wire.Value = &encoded
	}
	return wire
}

// encodeFailure projects err publicly and states the commit status by the
// contract's kind rule. write is false for read, list and resolve.
func encodeFailure(err error, write bool) *Failure {
	view, _ := faults.Public(err)
	failure := &Failure{
		Kind:    view.Kind.String(),
		Message: view.Message,
		Fields:  map[string]string{},
		Commit:  CommitNone,
	}
	if view.Type != "" {
		typ := view.Type
		failure.Type = &typ
	}
	for key, value := range view.Fields {
		failure.Fields[key] = value
	}
	if write {
		failure.Commit = CommitUnknown
		if view.Kind == faults.Invalid || view.Kind == faults.Conflict {
			failure.Commit = CommitNotApplied
		}
	}
	return failure
}
