package desktop

import (
	"strconv"
	"time"

	"github.com/0xsj/atelier-wails/internal/workspace/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

const (
	problemMissing = "missing"
	problemUnknown = "unknown"
	problemInvalid = "invalid"
)

func invalidRequest(path, problem string) error {
	return faults.New(faults.Invalid, "invalid request").WithType("desktop.invalid_request").WithField(path, problem)
}

func decodeID(path, value string) (id.ID, error) {
	if value == "" {
		return id.ID{}, invalidRequest(path, problemMissing)
	}
	parsed, err := id.Parse(value)
	if err != nil {
		return id.ID{}, invalidRequest(path, problemInvalid)
	}
	return parsed, nil
}

func decodeFilter(path, value string) (domain.ListFilter, error) {
	switch value {
	case FilterActive:
		return domain.ActiveOnly, nil
	case FilterAll:
		return domain.All, nil
	case "":
		return 0, invalidRequest(path, problemMissing)
	default:
		return 0, invalidRequest(path, problemUnknown)
	}
}

func decodeExpected(path string, value Expected) (domain.Expected, error) {
	switch value.Kind {
	case ExpectedAbsent:
		return domain.ExpectAbsent(), nil
	case ExpectedRevision:
		if value.Revision == nil {
			return domain.Expected{}, invalidRequest(path+".revision", problemMissing)
		}
		n, err := strconv.ParseUint(*value.Revision, 10, 64)
		if err != nil || n == 0 {
			return domain.Expected{}, invalidRequest(path+".revision", problemInvalid)
		}
		return domain.ExpectRevision(n)
	case "":
		return domain.Expected{}, invalidRequest(path+".kind", problemMissing)
	default:
		return domain.Expected{}, invalidRequest(path+".kind", problemUnknown)
	}
}

func decodeTime(path, value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, invalidRequest(path, problemInvalid)
	}
	return parsed.UTC().Truncate(time.Millisecond), nil
}

func encodeWorkspace(value domain.Workspace) Workspace {
	return Workspace{ID: value.ID().String(), Name: value.Name(), Location: value.Location(), Status: value.Status().String(), Revision: strconv.FormatUint(value.Revision(), 10), CreatedAt: formatTime(value.CreatedAt()), UpdatedAt: formatTime(value.UpdatedAt())}
}

func encodeEvent(value *domain.Event) *Event {
	if value == nil {
		return nil
	}
	event := &Event{Name: value.Name(), WorkspaceID: value.Workspace().String(), Revision: strconv.FormatUint(value.Revision(), 10), Title: value.Title(), Location: value.Location(), Status: value.Status().String()}
	return event
}

func encodeMutation(value domain.MutationResult) *MutationResponse {
	return &MutationResponse{Status: map[domain.ChangeStatus]string{domain.Changed: StatusChanged, domain.Unchanged: StatusUnchanged}[value.Status], Workspace: encodeWorkspace(value.Workspace), Event: encodeEvent(value.Event)}
}

func encodeFailure(err error, write bool) *Failure {
	view, _ := faults.Public(err)
	failure := &Failure{Kind: view.Kind.String(), Message: view.Message, Fields: map[string]string{}, Commit: CommitNone}
	if view.Type != "" {
		value := view.Type
		failure.Type = &value
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

func invalidDocument(write bool) *Failure {
	return encodeFailure(invalidRequest("request", problemInvalid), write)
}

func formatTime(value time.Time) string {
	return value.UTC().Truncate(time.Millisecond).Format("2006-01-02T15:04:05.000Z")
}
