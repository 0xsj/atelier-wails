// Package persistence stores the complete Workspace registry in one JSON
// document replaced atomically through fileio. A missing file is empty;
// corrupt state refuses startup. See CONTRACT.md.
package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/0xsj/atelier-wails/internal/workspace/domain"
	faults "github.com/0xsj/atelier-wails/pkg/errors"
	"github.com/0xsj/atelier-wails/pkg/fileio"
	"github.com/0xsj/atelier-wails/pkg/id"
)

const formatVersion = 1

var errUnsupported = errors.New("unsupported format")

type document struct {
	Format     int      `json:"format"`
	Workspaces []record `json:"workspaces"`
}

type record struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Location  string `json:"location"`
	Status    string `json:"status"`
	Revision  string `json:"revision"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Store is one file-backed Workspace registry. Reads are served from memory;
// a changed decision replaces the document before memory is updated.
type Store struct {
	mu    sync.RWMutex
	path  string
	items map[id.ID]domain.Workspace
}

// Open reads path once. Missing state is empty; invalid state is classified
// and returned rather than being silently discarded.
func Open(path string) (*Store, error) {
	data, found, err := fileio.Read(path)
	if err != nil {
		return nil, unavailable(path, err)
	}
	items := make(map[id.ID]domain.Workspace)
	if found {
		items, err = decodeDocument(data)
		if err != nil {
			if errors.Is(err, errUnsupported) {
				return nil, faults.New(faults.Unavailable, "workspace storage format is not supported").
					WithType("workspace.storage_unsupported").WithDetail("path", path)
			}
			return nil, faults.New(faults.Unavailable, "workspace storage is corrupt").
				WithType("workspace.storage_corrupt").
				WithDetail("path", path).
				WithDetail("reason", err.Error())
		}
	}
	return &Store{path: path, items: items}, nil
}

func (s *Store) Register(ctx context.Context, workspaceID id.ID, name, location string, at time.Time) (domain.RegisterResult, error) {
	if err := canceled(ctx); err != nil {
		return domain.RegisterResult{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[workspaceID]; exists {
		return domain.RegisterResult{}, conflict("workspace.id_taken")
	}
	if activeLocationTaken(s.items, location, workspaceID) {
		return domain.RegisterResult{}, conflict("workspace.location_taken")
	}
	result, err := domain.Register(workspaceID, name, location, at)
	if err != nil {
		return domain.RegisterResult{}, err
	}
	next := cloneItems(s.items)
	next[workspaceID] = result.Workspace
	if err := s.persist(next); err != nil {
		return domain.RegisterResult{}, err
	}
	s.items = next
	return result, nil
}

func (s *Store) Read(ctx context.Context, workspaceID id.ID) (domain.Workspace, bool, error) {
	if err := canceled(ctx); err != nil {
		return domain.Workspace{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if workspaceID.IsZero() {
		return domain.Workspace{}, false, invalid("workspace.invalid_id")
	}
	workspace, found := s.items[workspaceID]
	return workspace, found, nil
}

func (s *Store) List(ctx context.Context, filter domain.ListFilter) ([]domain.Workspace, error) {
	if err := canceled(ctx); err != nil {
		return nil, err
	}
	if !filter.Valid() {
		return nil, invalid("workspace.invalid_filter")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	workspaces := make([]domain.Workspace, 0, len(s.items))
	for _, workspace := range s.items {
		if filter == domain.ActiveOnly && workspace.Status() != domain.Active {
			continue
		}
		workspaces = append(workspaces, workspace)
	}
	sort.Slice(workspaces, func(i, j int) bool { return workspaces[i].ID().String() < workspaces[j].ID().String() })
	return workspaces, nil
}

func (s *Store) Rename(ctx context.Context, workspaceID id.ID, name string, expected domain.Expected, at time.Time) (domain.MutationResult, bool, error) {
	if err := canceled(ctx); err != nil {
		return domain.MutationResult{}, false, err
	}
	return s.mutate(workspaceID, func(current domain.Workspace, _ map[id.ID]domain.Workspace) (domain.MutationResult, error) {
		return domain.Rename(current, name, expected, at)
	})
}

func (s *Store) Archive(ctx context.Context, workspaceID id.ID, expected domain.Expected, at time.Time) (domain.MutationResult, bool, error) {
	if err := canceled(ctx); err != nil {
		return domain.MutationResult{}, false, err
	}
	return s.mutate(workspaceID, func(current domain.Workspace, _ map[id.ID]domain.Workspace) (domain.MutationResult, error) {
		return domain.Archive(current, expected, at)
	})
}

func (s *Store) Restore(ctx context.Context, workspaceID id.ID, expected domain.Expected, at time.Time) (domain.MutationResult, bool, error) {
	if err := canceled(ctx); err != nil {
		return domain.MutationResult{}, false, err
	}
	return s.mutate(workspaceID, func(current domain.Workspace, items map[id.ID]domain.Workspace) (domain.MutationResult, error) {
		if current.Status() == domain.Archived && activeLocationTaken(items, current.Location(), workspaceID) {
			return domain.MutationResult{}, conflict("workspace.location_taken")
		}
		return domain.Restore(current, expected, at)
	})
}

func (s *Store) Forget(ctx context.Context, workspaceID id.ID, expected domain.Expected) (domain.ForgetResult, bool, error) {
	if err := canceled(ctx); err != nil {
		return domain.ForgetResult{}, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, found := s.items[workspaceID]
	if !found {
		return domain.ForgetResult{}, false, nil
	}
	result, err := domain.Forget(current, expected)
	if err != nil {
		return domain.ForgetResult{}, true, err
	}
	next := cloneItems(s.items)
	delete(next, workspaceID)
	if err := s.persist(next); err != nil {
		return domain.ForgetResult{}, true, err
	}
	s.items = next
	return result, true, nil
}

func (s *Store) mutate(workspaceID id.ID, decide func(domain.Workspace, map[id.ID]domain.Workspace) (domain.MutationResult, error)) (domain.MutationResult, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, found := s.items[workspaceID]
	if !found {
		return domain.MutationResult{}, false, nil
	}
	result, err := decide(current, s.items)
	if err != nil || result.Status != domain.Changed {
		return result, true, err
	}
	next := cloneItems(s.items)
	next[workspaceID] = result.Workspace
	if err := s.persist(next); err != nil {
		return domain.MutationResult{}, true, err
	}
	s.items = next
	return result, true, nil
}

func (s *Store) persist(items map[id.ID]domain.Workspace) error {
	data, err := encodeDocument(items)
	if err != nil {
		return unavailable(s.path, err)
	}
	if err := fileio.Replace(s.path, data); err != nil {
		return unavailable(s.path, err)
	}
	return nil
}

func cloneItems(items map[id.ID]domain.Workspace) map[id.ID]domain.Workspace {
	next := make(map[id.ID]domain.Workspace, len(items))
	for key, value := range items {
		next[key] = value
	}
	return next
}

func activeLocationTaken(items map[id.ID]domain.Workspace, location string, except id.ID) bool {
	for otherID, workspace := range items {
		if otherID != except && workspace.Status() == domain.Active && workspace.Location() == location {
			return true
		}
	}
	return false
}

func encodeDocument(items map[id.ID]domain.Workspace) ([]byte, error) {
	records := make([]record, 0, len(items))
	for _, workspace := range items {
		records = append(records, encodeRecord(workspace))
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	data, err := json.Marshal(document{Format: formatVersion, Workspaces: records})
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func encodeRecord(workspace domain.Workspace) record {
	return record{
		ID: workspace.ID().String(), Name: workspace.Name(), Location: workspace.Location(),
		Status: workspace.Status().String(), Revision: strconv.FormatUint(workspace.Revision(), 10),
		CreatedAt: strconv.FormatInt(workspace.CreatedAt().UnixMilli(), 10),
		UpdatedAt: strconv.FormatInt(workspace.UpdatedAt().UnixMilli(), 10),
	}
}

func decodeDocument(data []byte) (map[id.ID]domain.Workspace, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var stored document
	if err := decoder.Decode(&stored); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, errors.New("trailing JSON")
		}
		return nil, err
	}
	if stored.Format != formatVersion {
		return nil, errUnsupported
	}
	items := make(map[id.ID]domain.Workspace, len(stored.Workspaces))
	for index, value := range stored.Workspaces {
		workspace, err := decodeRecord(value)
		if err != nil {
			return nil, errors.New("workspace " + strconv.Itoa(index) + ": " + err.Error())
		}
		if activeLocationTaken(items, workspace.Location(), workspace.ID()) {
			return nil, errors.New("duplicate active location")
		}
		if _, exists := items[workspace.ID()]; exists {
			return nil, errors.New("duplicate id")
		}
		items[workspace.ID()] = workspace
	}
	return items, nil
}

func decodeRecord(value record) (domain.Workspace, error) {
	workspaceID, err := id.Parse(value.ID)
	if err != nil {
		return domain.Workspace{}, errors.New("id")
	}
	status := map[string]domain.Status{"active": domain.Active, "archived": domain.Archived}[value.Status]
	if !status.Valid() {
		return domain.Workspace{}, errors.New("status")
	}
	revision, err := strconv.ParseUint(value.Revision, 10, 64)
	if err != nil || revision == 0 {
		return domain.Workspace{}, errors.New("revision")
	}
	createdAt, err := parseMillis(value.CreatedAt)
	if err != nil {
		return domain.Workspace{}, errors.New("created_at")
	}
	updatedAt, err := parseMillis(value.UpdatedAt)
	if err != nil {
		return domain.Workspace{}, errors.New("updated_at")
	}
	workspace, err := domain.Rebuild(workspaceID, value.Name, value.Location, status, revision, createdAt, updatedAt)
	if err != nil {
		return domain.Workspace{}, errors.New("workspace")
	}
	return workspace, nil
}

func parseMillis(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("empty timestamp")
	}
	valueMillis, err := strconv.ParseInt(value, 10, 64)
	if err != nil || valueMillis < 0 {
		return time.Time{}, errors.New("timestamp")
	}
	return time.UnixMilli(valueMillis).UTC(), nil
}

func invalid(typ string) error {
	return faults.New(faults.Invalid, "invalid workspace input").WithType(typ)
}

func conflict(typ string) error {
	return faults.New(faults.Conflict, "workspace already registered").WithType(typ)
}

func canceled(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return faults.Wrap(err, faults.Canceled, "workspace operation canceled")
	}
	return nil
}

func unavailable(path string, cause error) error {
	return faults.New(faults.Unavailable, "workspace storage is unavailable").
		WithType("workspace.storage_unavailable").WithDetail("path", path).WithCause(cause)
}
