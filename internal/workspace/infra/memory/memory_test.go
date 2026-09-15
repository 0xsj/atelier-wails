package memory_test

import (
	"testing"

	"github.com/0xsj/atelier-wails/internal/workspace/infra/memory"
	"github.com/0xsj/atelier-wails/internal/workspace/infra/storetest"
)

func TestWorkspaceMemoryContract(t *testing.T) {
	storetest.Run(t, func(*testing.T) storetest.Store { return memory.New() })
}
