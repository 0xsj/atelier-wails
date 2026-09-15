package root

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	faults "github.com/0xsj/atelier-wails/pkg/errors"
)

const createDocument = `{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"text","text":"dark"},"expected":{"kind":"absent"}}`

const createFixture = `{"ok":true,"value":{"status":"changed","entry":{"scope":{"kind":"global"},"key":"editor.theme","value":{"kind":"text","text":"dark"},"revision":"1"},"event":{"name":"preference.changed","scope":{"kind":"global"},"key":"editor.theme","revision":"1","value":{"kind":"text","text":"dark"}}}}`

const readDocument = `{"scope":{"kind":"global"},"key":"editor.theme"}`

func encode(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func documentPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "atelier-wails", "preferences.json")
}

// N01/N02: the composed facade serves the wire round trip over the
// file-backed store root selected, creating the config directory.
func TestComposedPreferencesRoundTrip(t *testing.T) {
	path := documentPath(t)
	facade, err := composePreferences(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := encode(t, facade.Replace(createDocument)); got != createFixture {
		t.Fatalf("create through facade:\n got %s\nwant %s", got, createFixture)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("document not written: %v", err)
	}
	read := facade.Read(readDocument)
	if !read.OK || !read.Value.Found || read.Value.Entry.Revision != "1" {
		t.Fatalf("read through facade: %+v", read)
	}
	listed := facade.List(`{"scope":{"kind":"global"}}`)
	if !listed.OK || len(listed.Value.Entries) != 1 {
		t.Fatalf("list through facade: %+v", listed)
	}
	resolved := facade.Resolve(`{"scope":{"kind":"global"},"key":"editor.wrap","fallback":{"kind":"bool","bool":true}}`)
	if !resolved.OK || resolved.Value.Stored {
		t.Fatalf("resolve through facade: %+v", resolved)
	}
	removed := facade.Remove(`{"scope":{"kind":"global"},"key":"editor.theme","expected":{"kind":"revision","revision":"1"}}`)
	if !removed.OK || !removed.Value.Removed || *removed.Value.Revision != "2" {
		t.Fatalf("remove through facade: %+v", removed)
	}
}

// N03: a document the facade cannot decode is the contract's failure.
func TestComposedPreferencesInvalidDocument(t *testing.T) {
	facade, err := composePreferences(documentPath(t))
	if err != nil {
		t.Fatal(err)
	}
	write := facade.Replace(`not json`)
	if got := encode(t, write); got != `{"ok":false,"failure":{"kind":"invalid","message":"invalid request","type":"desktop.invalid_request","fields":{"request":"invalid"},"commit":"not_applied"}}` {
		t.Fatalf("invalid write document: %s", got)
	}
	read := facade.Read(`not json`)
	if read.OK || read.Failure == nil || read.Failure.Commit != "none" {
		t.Fatalf("invalid read document: %+v", read)
	}
}

// A completed write survives a recomposition over the same path.
func TestComposedPreferencesSurviveRestart(t *testing.T) {
	path := documentPath(t)
	first, err := composePreferences(path)
	if err != nil {
		t.Fatal(err)
	}
	if outcome := first.Replace(createDocument); !outcome.OK {
		t.Fatalf("create: %+v", outcome.Failure)
	}
	again, err := composePreferences(path)
	if err != nil {
		t.Fatal(err)
	}
	read := again.Read(readDocument)
	if !read.OK || !read.Value.Found || read.Value.Entry.Revision != "1" || *read.Value.Entry.Value.Text != "dark" {
		t.Fatalf("entry did not survive restart: %+v", read)
	}
}

// A corrupt document refuses composition instead of starting empty.
func TestCorruptDocumentRefusesComposition(t *testing.T) {
	path := documentPath(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	facade, err := composePreferences(path)
	if facade != nil || !faults.IsKind(err, faults.Unavailable) || faults.DiagnosticTypeOf(err) != "preferences.storage_corrupt" {
		t.Fatalf("corrupt document: %v %v", facade, err)
	}
}

// N05: each composition over its own path owns its own state.
func TestComposedPreferencesAreIndependent(t *testing.T) {
	first, err := composePreferences(documentPath(t))
	if err != nil {
		t.Fatal(err)
	}
	second, err := composePreferences(documentPath(t))
	if err != nil {
		t.Fatal(err)
	}
	if outcome := first.Replace(createDocument); !outcome.OK {
		t.Fatalf("create in first: %+v", outcome.Failure)
	}
	if read := second.Read(readDocument); !read.OK || read.Value.Found {
		t.Fatalf("second composition shares state: %+v", read)
	}
}
