package desktop

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/0xsj/atelier-wails/internal/workspace/app/command"
	"github.com/0xsj/atelier-wails/internal/workspace/app/query"
	"github.com/0xsj/atelier-wails/internal/workspace/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/id"
)

// Handler turns decoded wire requests into outcome envelopes. It has no
// framework dependency; the Wails host facade owns binding only.
type Handler struct {
	commands *command.Service
	queries  *query.Service
}

func NewHandler(commands *command.Service, queries *query.Service) (*Handler, error) {
	if commands == nil || queries == nil {
		return nil, missingDependency()
	}
	return &Handler{commands: commands, queries: queries}, nil
}

func (h *Handler) Read(ctx context.Context, request ReadRequest) ReadOutcome {
	workspaceID, err := decodeID("id", request.ID)
	if err != nil {
		return ReadOutcome{Failure: encodeFailure(err, false)}
	}
	value, found, err := h.queries.Read(ctx, workspaceID)
	if err != nil {
		return ReadOutcome{Failure: encodeFailure(err, false)}
	}
	response := &ReadResponse{Found: found}
	if found {
		encoded := encodeWorkspace(value)
		response.Workspace = &encoded
	}
	return ReadOutcome{OK: true, Value: response}
}

func (h *Handler) List(ctx context.Context, request ListRequest) ListOutcome {
	filter, err := decodeFilter("filter", request.Filter)
	if err != nil {
		return ListOutcome{Failure: encodeFailure(err, false)}
	}
	values, err := h.queries.List(ctx, filter)
	if err != nil {
		return ListOutcome{Failure: encodeFailure(err, false)}
	}
	response := &ListResponse{Workspaces: make([]Workspace, 0, len(values))}
	for _, value := range values {
		response.Workspaces = append(response.Workspaces, encodeWorkspace(value))
	}
	return ListOutcome{OK: true, Value: response}
}

func (h *Handler) Register(ctx context.Context, request RegisterRequest) RegisteredOutcome {
	workspaceID, err := decodeID("id", request.ID)
	if err != nil {
		return RegisteredOutcome{Failure: encodeFailure(err, true)}
	}
	at, err := decodeTime("at", request.At)
	if err != nil {
		return RegisteredOutcome{Failure: encodeFailure(err, true)}
	}
	result, err := h.commands.Register(ctx, workspaceID, request.Name, request.Location, at)
	if err != nil {
		return RegisteredOutcome{Failure: encodeFailure(err, true)}
	}
	return RegisteredOutcome{OK: true, Value: &RegisteredResponse{Workspace: encodeWorkspace(result.Workspace), Event: *encodeEvent(&result.Event)}}
}

func (h *Handler) Rename(ctx context.Context, request RenameRequest) MutationOutcome {
	workspaceID, expected, at, err := decodeMutation(request.ID, request.Expected, request.At)
	if err != nil {
		return MutationOutcome{Failure: encodeFailure(err, true)}
	}
	result, err := h.commands.Rename(ctx, workspaceID, request.Name, expected, at)
	if err != nil {
		return MutationOutcome{Failure: encodeFailure(err, true)}
	}
	return MutationOutcome{OK: true, Value: encodeMutation(result)}
}

func (h *Handler) Archive(ctx context.Context, request LifecycleRequest) MutationOutcome {
	workspaceID, expected, at, err := decodeLifecycle(request)
	if err != nil {
		return MutationOutcome{Failure: encodeFailure(err, true)}
	}
	result, err := h.commands.Archive(ctx, workspaceID, expected, at)
	if err != nil {
		return MutationOutcome{Failure: encodeFailure(err, true)}
	}
	return MutationOutcome{OK: true, Value: encodeMutation(result)}
}

func (h *Handler) Restore(ctx context.Context, request LifecycleRequest) MutationOutcome {
	workspaceID, expected, at, err := decodeLifecycle(request)
	if err != nil {
		return MutationOutcome{Failure: encodeFailure(err, true)}
	}
	result, err := h.commands.Restore(ctx, workspaceID, expected, at)
	if err != nil {
		return MutationOutcome{Failure: encodeFailure(err, true)}
	}
	return MutationOutcome{OK: true, Value: encodeMutation(result)}
}

func (h *Handler) Forget(ctx context.Context, request LifecycleRequest) ForgetOutcome {
	workspaceID, err := decodeID("id", request.ID)
	if err != nil {
		return ForgetOutcome{Failure: encodeFailure(err, true)}
	}
	expected, err := decodeExpected("expected", request.Expected)
	if err != nil {
		return ForgetOutcome{Failure: encodeFailure(err, true)}
	}
	result, found, err := h.commands.Forget(ctx, workspaceID, expected)
	if err != nil {
		return ForgetOutcome{Failure: encodeFailure(err, true)}
	}
	response := &ForgetResponse{Found: found}
	if found {
		response.Workspace = &ForgottenResponse{Revision: strconv.FormatUint(result.Revision, 10), Event: *encodeEvent(&result.Event)}
	}
	return ForgetOutcome{OK: true, Value: response}
}

func decodeMutation(idText string, expectedWire Expected, atText string) (id.ID, domain.Expected, time.Time, error) {
	workspaceID, err := decodeID("id", idText)
	if err != nil {
		return id.ID{}, domain.Expected{}, time.Time{}, err
	}
	expected, err := decodeExpected("expected", expectedWire)
	if err != nil {
		return id.ID{}, domain.Expected{}, time.Time{}, err
	}
	at, err := decodeTime("at", atText)
	return workspaceID, expected, at, err
}

func decodeLifecycle(request LifecycleRequest) (id.ID, domain.Expected, time.Time, error) {
	if request.At == nil {
		return id.ID{}, domain.Expected{}, time.Time{}, invalidRequest("at", problemMissing)
	}
	return decodeMutation(request.ID, request.Expected, *request.At)
}

func (h *Handler) ReadDocument(ctx context.Context, document []byte) ReadOutcome {
	var request ReadRequest
	if !decodeDocument(document, &request) {
		return ReadOutcome{Failure: invalidDocument(false)}
	}
	return h.Read(ctx, request)
}
func (h *Handler) ListDocument(ctx context.Context, document []byte) ListOutcome {
	var request ListRequest
	if !decodeDocument(document, &request) {
		return ListOutcome{Failure: invalidDocument(false)}
	}
	return h.List(ctx, request)
}
func (h *Handler) RegisterDocument(ctx context.Context, document []byte) RegisteredOutcome {
	var request RegisterRequest
	if !decodeDocument(document, &request) {
		return RegisteredOutcome{Failure: invalidDocument(true)}
	}
	return h.Register(ctx, request)
}
func (h *Handler) RenameDocument(ctx context.Context, document []byte) MutationOutcome {
	var request RenameRequest
	if !decodeDocument(document, &request) {
		return MutationOutcome{Failure: invalidDocument(true)}
	}
	return h.Rename(ctx, request)
}
func (h *Handler) ArchiveDocument(ctx context.Context, document []byte) MutationOutcome {
	var request LifecycleRequest
	if !decodeDocument(document, &request) {
		return MutationOutcome{Failure: invalidDocument(true)}
	}
	return h.Archive(ctx, request)
}
func (h *Handler) RestoreDocument(ctx context.Context, document []byte) MutationOutcome {
	var request LifecycleRequest
	if !decodeDocument(document, &request) {
		return MutationOutcome{Failure: invalidDocument(true)}
	}
	return h.Restore(ctx, request)
}
func (h *Handler) ForgetDocument(ctx context.Context, document []byte) ForgetOutcome {
	var request LifecycleRequest
	if !decodeDocument(document, &request) {
		return ForgetOutcome{Failure: invalidDocument(true)}
	}
	return h.Forget(ctx, request)
}

func decodeDocument(document []byte, into any) bool {
	trimmed := bytes.TrimSpace(document)
	return len(trimmed) > 0 && trimmed[0] == '{' && json.Unmarshal(trimmed, into) == nil
}

func missingDependency() error {
	return faults.New(faults.Internal, "workspace desktop handler needs both services").WithType("workspace.missing_dependency")
}
