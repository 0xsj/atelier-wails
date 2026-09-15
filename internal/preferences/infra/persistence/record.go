package persistence

import (
	"bytes"
	"encoding/json"
	"errors"
	"sort"
	"strconv"

	"github.com/0xsj/atelier-wails/internal/preferences/domain"
	"github.com/0xsj/atelier-wails/pkg/id"
)

// formatVersion is the only document format this revision reads or writes.
const formatVersion = 1

// The records below are persistence-owned. They mirror the wire shapes
// today but are a separate schema.

type document struct {
	Format  int      `json:"format"`
	Entries []record `json:"entries"`
}

type record struct {
	Scope    scopeRecord `json:"scope"`
	Key      string      `json:"key"`
	Value    valueRecord `json:"value"`
	Revision string      `json:"revision"`
}

type scopeRecord struct {
	Kind        string `json:"kind"`
	WorkspaceID string `json:"workspace_id,omitempty"`
}

type valueRecord struct {
	Kind string  `json:"kind"`
	Text *string `json:"text,omitempty"`
	Bool *bool   `json:"bool,omitempty"`
	Int  *string `json:"int,omitempty"`
}

type items map[domain.Scope]map[domain.Key]domain.Entry

// encodeDocument renders items sorted by scope then key, so identical state
// produces identical bytes.
func encodeDocument(state items) ([]byte, error) {
	entries := make([]record, 0)
	scopes := make([]domain.Scope, 0, len(state))
	for scope := range state {
		scopes = append(scopes, scope)
	}
	sort.Slice(scopes, func(i, j int) bool { return scopes[i].String() < scopes[j].String() })
	for _, scope := range scopes {
		keys := make([]domain.Key, 0, len(state[scope]))
		for key := range state[scope] {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
		for _, key := range keys {
			entries = append(entries, encodeRecord(state[scope][key]))
		}
	}
	data, err := json.Marshal(document{Format: formatVersion, Entries: entries})
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func encodeRecord(entry domain.Entry) record {
	out := record{Key: entry.Key().String(), Revision: strconv.FormatUint(entry.Revision(), 10)}
	if workspace, ok := entry.Scope().WorkspaceID(); ok {
		out.Scope = scopeRecord{Kind: "workspace", WorkspaceID: workspace.String()}
	} else {
		out.Scope = scopeRecord{Kind: "global"}
	}
	switch value := entry.Value(); value.Kind() {
	case domain.TextValue:
		text, _ := value.Text()
		out.Value = valueRecord{Kind: "text", Text: &text}
	case domain.BoolValue:
		flag, _ := value.Bool()
		out.Value = valueRecord{Kind: "bool", Bool: &flag}
	default:
		n, _ := value.Int()
		text := strconv.FormatInt(n, 10)
		out.Value = valueRecord{Kind: "int", Int: &text}
	}
	return out
}

// errUnsupported marks a document with a format other than formatVersion.
var errUnsupported = errors.New("unsupported format")

// decodeDocument restores items through validated domain construction. Any
// structural or domain refusal is corruption; only the format mismatch is
// reported separately.
func decodeDocument(data []byte) (items, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var doc document
	if err := decoder.Decode(&doc); err != nil {
		return nil, err
	}
	if decoder.More() {
		return nil, errors.New("trailing content after document")
	}
	if doc.Format != formatVersion {
		return nil, errUnsupported
	}
	state := make(items)
	for index, rec := range doc.Entries {
		entry, err := decodeRecord(rec)
		if err != nil {
			return nil, errors.New("entry " + strconv.Itoa(index) + ": " + err.Error())
		}
		bucket := state[entry.Scope()]
		if bucket == nil {
			bucket = make(map[domain.Key]domain.Entry)
			state[entry.Scope()] = bucket
		}
		if _, duplicate := bucket[entry.Key()]; duplicate {
			return nil, errors.New("entry " + strconv.Itoa(index) + ": duplicate scope and key")
		}
		bucket[entry.Key()] = entry
	}
	return state, nil
}

func decodeRecord(rec record) (domain.Entry, error) {
	var scope domain.Scope
	switch rec.Scope.Kind {
	case "global":
		scope = domain.Global()
	case "workspace":
		parsed, err := id.Parse(rec.Scope.WorkspaceID)
		if err != nil {
			return domain.Entry{}, errors.New("workspace id")
		}
		scope, err = domain.ForWorkspace(parsed)
		if err != nil {
			return domain.Entry{}, errors.New("workspace scope")
		}
	default:
		return domain.Entry{}, errors.New("scope kind")
	}
	key, err := domain.NewKey(rec.Key)
	if err != nil {
		return domain.Entry{}, errors.New("key")
	}
	var value domain.Value
	switch rec.Value.Kind {
	case "text":
		if rec.Value.Text == nil {
			return domain.Entry{}, errors.New("text")
		}
		if value, err = domain.Text(*rec.Value.Text); err != nil {
			return domain.Entry{}, errors.New("text")
		}
	case "bool":
		if rec.Value.Bool == nil {
			return domain.Entry{}, errors.New("bool")
		}
		value = domain.Bool(*rec.Value.Bool)
	case "int":
		if rec.Value.Int == nil {
			return domain.Entry{}, errors.New("int")
		}
		n, err := strconv.ParseInt(*rec.Value.Int, 10, 64)
		if err != nil {
			return domain.Entry{}, errors.New("int")
		}
		value = domain.Int(n)
	default:
		return domain.Entry{}, errors.New("value kind")
	}
	revision, err := strconv.ParseUint(rec.Revision, 10, 64)
	if err != nil {
		return domain.Entry{}, errors.New("revision")
	}
	entry, err := domain.Restore(scope, key, value, revision)
	if err != nil {
		return domain.Entry{}, errors.New("revision")
	}
	return entry, nil
}
