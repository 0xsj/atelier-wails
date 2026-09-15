package desktop

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/0xsj/atelier-wails/internal/workspace/app/command"
	"github.com/0xsj/atelier-wails/internal/workspace/app/query"
	"github.com/0xsj/atelier-wails/internal/workspace/infra/memory"
)

const (
	testID      = "01900000-0000-7000-8000-000000000001"
	testOtherID = "01900000-0000-7000-8000-000000000002"
	testAt      = "2026-09-15T12:00:00.123Z"
	testLater   = "2026-09-15T12:00:01.456Z"
)

func testHandler(t *testing.T, store *memory.Store) *Handler {
	t.Helper()
	commands, err := command.New(store, command.Discard{})
	if err != nil {
		t.Fatal(err)
	}
	queries, err := query.New(store)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(commands, queries)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func testDocument(t *testing.T, value any) []byte {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func testRevision(value string) Expected {
	return Expected{Kind: ExpectedRevision, Revision: &value}
}

func registerRequest() RegisterRequest {
	return RegisterRequest{ID: testID, Name: "Atelier", Location: "/tmp/atelier", At: testAt}
}

func TestWT01DocumentsRequireObjects(t *testing.T) {
	handler := testHandler(t, memory.New())

	read := handler.ReadDocument(context.Background(), []byte("null"))
	if read.Failure == nil || read.Failure.Type == nil || *read.Failure.Type != "desktop.invalid_request" {
		t.Fatalf("unexpected read failure: %+v", read.Failure)
	}
	if read.Failure.Commit != CommitNone || read.Failure.Fields["request"] != problemInvalid {
		t.Fatalf("unexpected read projection: %+v", read.Failure)
	}

	register := handler.RegisterDocument(context.Background(), []byte("[]"))
	if register.Failure == nil || register.Failure.Commit != CommitNotApplied {
		t.Fatalf("unexpected register failure: %+v", register.Failure)
	}
}

func TestWT02RegisterReadAndList(t *testing.T) {
	handler := testHandler(t, memory.New())
	ctx := context.Background()

	registered := handler.RegisterDocument(ctx, testDocument(t, registerRequest()))
	if !registered.OK || registered.Value == nil {
		t.Fatalf("register failed: %+v", registered)
	}
	if registered.Value.Workspace.Revision != "1" || registered.Value.Event.Name != "workspace.registered" {
		t.Fatalf("unexpected register response: %+v", registered.Value)
	}

	read := handler.ReadDocument(ctx, testDocument(t, ReadRequest{ID: testID}))
	if !read.OK || read.Value == nil || !read.Value.Found || read.Value.Workspace == nil {
		t.Fatalf("read failed: %+v", read)
	}
	if read.Value.Workspace.Location != "/tmp/atelier" {
		t.Fatalf("unexpected read response: %+v", read.Value.Workspace)
	}

	list := handler.ListDocument(ctx, testDocument(t, ListRequest{Filter: FilterActive}))
	if !list.OK || list.Value == nil || len(list.Value.Workspaces) != 1 {
		t.Fatalf("list failed: %+v", list)
	}
}

func TestWT03LifecycleResponsesMirrorEvents(t *testing.T) {
	handler := testHandler(t, memory.New())
	ctx := context.Background()
	register := handler.Register(ctx, registerRequest())
	if !register.OK || register.Value == nil {
		t.Fatalf("register failed: %+v", register)
	}

	rename := handler.Rename(ctx, RenameRequest{ID: testID, Name: "North", Expected: testRevision("1"), At: testLater})
	if !rename.OK || rename.Value == nil || rename.Value.Status != StatusChanged || rename.Value.Event == nil {
		t.Fatalf("rename failed: %+v", rename)
	}
	if rename.Value.Event.Name != "workspace.renamed" || rename.Value.Event.Revision != "2" {
		t.Fatalf("unexpected rename event: %+v", rename.Value.Event)
	}

	archive := handler.Archive(ctx, LifecycleRequest{ID: testID, Expected: testRevision("2"), At: stringPtr(testLater)})
	if !archive.OK || archive.Value == nil || archive.Value.Event == nil || archive.Value.Event.Name != "workspace.archived" {
		t.Fatalf("archive failed: %+v", archive)
	}
	restore := handler.Restore(ctx, LifecycleRequest{ID: testID, Expected: testRevision("3"), At: stringPtr(testLater)})
	if !restore.OK || restore.Value == nil || restore.Value.Event == nil || restore.Value.Event.Name != "workspace.restored" {
		t.Fatalf("restore failed: %+v", restore)
	}

	archive = handler.Archive(ctx, LifecycleRequest{ID: testID, Expected: testRevision("4"), At: stringPtr(testLater)})
	forget := handler.Forget(ctx, LifecycleRequest{ID: testID, Expected: testRevision("5")})
	if !archive.OK || !forget.OK || forget.Value == nil || !forget.Value.Found || forget.Value.Workspace == nil {
		t.Fatalf("forget flow failed: archive=%+v forget=%+v", archive, forget)
	}
	if forget.Value.Workspace.Event.Name != "workspace.forgotten" || forget.Value.Workspace.Revision != "6" {
		t.Fatalf("unexpected forget response: %+v", forget.Value.Workspace)
	}
}

func TestWT04RevisionsAndTimestampsAreCanonical(t *testing.T) {
	handler := testHandler(t, memory.New())
	ctx := context.Background()
	registered := handler.Register(ctx, registerRequest())
	if registered.Value == nil {
		t.Fatalf("register failed: %+v", registered)
	}
	encoded, err := json.Marshal(registered.Value.Workspace)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"revision":"1"`) || !strings.Contains(string(encoded), testAt) {
		t.Fatalf("non-canonical workspace encoding: %s", encoded)
	}

	bad := handler.Rename(ctx, RenameRequest{ID: testID, Name: "North", Expected: testRevision("0"), At: testLater})
	if bad.Failure == nil || bad.Failure.Fields["expected.revision"] != problemInvalid {
		t.Fatalf("unexpected invalid revision response: %+v", bad.Failure)
	}
}

func TestWT05InvalidFieldsDoNotEchoInput(t *testing.T) {
	handler := testHandler(t, memory.New())
	secret := "not-a-real-workspace-id"
	outcome := handler.Read(context.Background(), ReadRequest{ID: secret})
	if outcome.Failure == nil || outcome.Failure.Fields["id"] != problemInvalid {
		t.Fatalf("unexpected invalid id response: %+v", outcome.Failure)
	}
	encoded, err := json.Marshal(outcome)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), secret) {
		t.Fatalf("failure echoed untrusted input: %s", encoded)
	}
}

func TestWT06FailuresPreserveClassificationAndCommit(t *testing.T) {
	handler := testHandler(t, memory.New())
	ctx := context.Background()
	if result := handler.Register(ctx, registerRequest()); !result.OK {
		t.Fatalf("initial register failed: %+v", result)
	}
	duplicate := handler.Register(ctx, registerRequest())
	if duplicate.Failure == nil || duplicate.Failure.Kind != "conflict" || duplicate.Failure.Commit != CommitNotApplied {
		t.Fatalf("unexpected duplicate projection: %+v", duplicate.Failure)
	}
	if duplicate.Failure.Type == nil || *duplicate.Failure.Type != "workspace.id_taken" {
		t.Fatalf("unexpected duplicate type: %+v", duplicate.Failure)
	}
}

func TestWT07OneHandlerServesSharedStoreState(t *testing.T) {
	store := memory.New()
	first := testHandler(t, store)
	second := testHandler(t, store)
	if result := first.Register(context.Background(), registerRequest()); !result.OK {
		t.Fatalf("register failed: %+v", result)
	}
	read := second.Read(context.Background(), ReadRequest{ID: testID})
	if !read.OK || read.Value == nil || !read.Value.Found || read.Value.Workspace == nil {
		t.Fatalf("shared store read failed: %+v", read)
	}
	if read.Value.Workspace.ID != testID {
		t.Fatalf("unexpected shared workspace: %+v", read.Value.Workspace)
	}

	other := second.Read(context.Background(), ReadRequest{ID: testOtherID})
	if !other.OK || other.Value == nil || other.Value.Found {
		t.Fatalf("unexpected absent workspace response: %+v", other)
	}
}

func stringPtr(value string) *string { return &value }
