package desktop

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/0xsj/atelier-wails/internal/preferences/app/command"
	"github.com/0xsj/atelier-wails/internal/preferences/app/query"
	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
)

// Handler turns decoded wire requests into outcome envelopes. The native
// facade owns JSON decoding of the request document and forwards here.
type Handler struct {
	commands *command.Service
	queries  *query.Service
}

// NewHandler refuses nil services.
func NewHandler(commands *command.Service, queries *query.Service) (*Handler, error) {
	if commands == nil || queries == nil {
		return nil, faults.New(faults.Internal, "preferences desktop handler needs both services").
			WithType("preferences.missing_dependency")
	}
	return &Handler{commands: commands, queries: queries}, nil
}

func (h *Handler) Read(ctx context.Context, request ReadRequest) ReadOutcome {
	scope, err := decodeScope("scope", request.Scope)
	if err != nil {
		return ReadOutcome{Failure: encodeFailure(err, false)}
	}
	key, err := decodeKey(request.Key)
	if err != nil {
		return ReadOutcome{Failure: encodeFailure(err, false)}
	}
	entry, found, err := h.queries.Read(ctx, scope, key)
	if err != nil {
		return ReadOutcome{Failure: encodeFailure(err, false)}
	}
	response := &ReadResponse{Found: found}
	if found {
		encoded := encodeEntry(entry)
		response.Entry = &encoded
	}
	return ReadOutcome{OK: true, Value: response}
}

func (h *Handler) List(ctx context.Context, request ListRequest) ListOutcome {
	scope, err := decodeScope("scope", request.Scope)
	if err != nil {
		return ListOutcome{Failure: encodeFailure(err, false)}
	}
	entries, err := h.queries.List(ctx, scope)
	if err != nil {
		return ListOutcome{Failure: encodeFailure(err, false)}
	}
	response := &ListResponse{Entries: make([]Entry, 0, len(entries))}
	for _, entry := range entries {
		response.Entries = append(response.Entries, encodeEntry(entry))
	}
	return ListOutcome{OK: true, Value: response}
}

func (h *Handler) Replace(ctx context.Context, request ReplaceRequest) ReplaceOutcome {
	scope, err := decodeScope("scope", request.Scope)
	if err != nil {
		return ReplaceOutcome{Failure: encodeFailure(err, true)}
	}
	key, err := decodeKey(request.Key)
	if err != nil {
		return ReplaceOutcome{Failure: encodeFailure(err, true)}
	}
	value, err := decodeValue("value", request.Value)
	if err != nil {
		return ReplaceOutcome{Failure: encodeFailure(err, true)}
	}
	expected, err := decodeExpected("expected", request.Expected)
	if err != nil {
		return ReplaceOutcome{Failure: encodeFailure(err, true)}
	}
	result, err := h.commands.Replace(ctx, scope, key, value, expected)
	if err != nil {
		return ReplaceOutcome{Failure: encodeFailure(err, true)}
	}
	status := StatusSame
	if result.Status == domain.Changed {
		status = StatusChanged
	}
	return ReplaceOutcome{OK: true, Value: &ReplaceResponse{
		Status: status,
		Entry:  encodeEntry(result.Entry),
		Event:  encodeEvent(result.Event),
	}}
}

func (h *Handler) Remove(ctx context.Context, request RemoveRequest) RemoveOutcome {
	scope, err := decodeScope("scope", request.Scope)
	if err != nil {
		return RemoveOutcome{Failure: encodeFailure(err, true)}
	}
	key, err := decodeKey(request.Key)
	if err != nil {
		return RemoveOutcome{Failure: encodeFailure(err, true)}
	}
	expected, err := decodeExpected("expected", request.Expected)
	if err != nil {
		return RemoveOutcome{Failure: encodeFailure(err, true)}
	}
	result, err := h.commands.Remove(ctx, scope, key, expected)
	if err != nil {
		return RemoveOutcome{Failure: encodeFailure(err, true)}
	}
	response := &RemoveResponse{Removed: result.Removed}
	if result.Removed {
		revision := encodeRevision(result.Revision)
		response.Revision = &revision
		response.Event = encodeEvent(result.Event)
	}
	return RemoveOutcome{OK: true, Value: response}
}

func (h *Handler) Resolve(ctx context.Context, request ResolveRequest) ResolveOutcome {
	scope, err := decodeScope("scope", request.Scope)
	if err != nil {
		return ResolveOutcome{Failure: encodeFailure(err, false)}
	}
	key, err := decodeKey(request.Key)
	if err != nil {
		return ResolveOutcome{Failure: encodeFailure(err, false)}
	}
	fallback, err := decodeValue("fallback", request.Fallback)
	if err != nil {
		return ResolveOutcome{Failure: encodeFailure(err, false)}
	}
	resolved, err := h.queries.Resolve(ctx, scope, key, fallback)
	if err != nil {
		return ResolveOutcome{Failure: encodeFailure(err, false)}
	}
	response := &ResolveResponse{Value: encodeValue(resolved.Value), Stored: resolved.Stored}
	if resolved.Stored {
		revision := encodeRevision(resolved.Revision)
		response.Revision = &revision
	}
	return ResolveOutcome{OK: true, Value: response}
}

// Document entry points decode a raw JSON request document before dispatch,
// so a native facade can accept the document as one string argument and the
// contract's decode failure is produced here rather than by the framework.

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

func (h *Handler) ReplaceDocument(ctx context.Context, document []byte) ReplaceOutcome {
	var request ReplaceRequest
	if !decodeDocument(document, &request) {
		return ReplaceOutcome{Failure: invalidDocument(true)}
	}
	return h.Replace(ctx, request)
}

func (h *Handler) RemoveDocument(ctx context.Context, document []byte) RemoveOutcome {
	var request RemoveRequest
	if !decodeDocument(document, &request) {
		return RemoveOutcome{Failure: invalidDocument(true)}
	}
	return h.Remove(ctx, request)
}

func (h *Handler) ResolveDocument(ctx context.Context, document []byte) ResolveOutcome {
	var request ResolveRequest
	if !decodeDocument(document, &request) {
		return ResolveOutcome{Failure: invalidDocument(false)}
	}
	return h.Resolve(ctx, request)
}

// decodeDocument requires a JSON object, as every request shape is one; a
// null or scalar the decoder would otherwise tolerate is refused.
func decodeDocument(document []byte, into any) bool {
	trimmed := bytes.TrimSpace(document)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return false
	}
	return json.Unmarshal(trimmed, into) == nil
}

// invalidDocument is the contract's failure for a document that does not
// decode into the request type: {"request":"invalid"}.
func invalidDocument(write bool) *Failure {
	return encodeFailure(invalidRequest("request", problemInvalid), write)
}
