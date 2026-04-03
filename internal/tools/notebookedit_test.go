package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotebookEditTool_Execute(t *testing.T) {
	// Create a dummy notebook
	nb := JupyterNotebook{
		Cells: []NotebookCell{
			{CellType: "markdown", Source: []string{"# Title"}},
			{CellType: "code", Source: []string{"print('hello')"}, Outputs: []interface{}{}},
		},
		Metadata:      make(map[string]interface{}),
		Nbformat:      4,
		NbformatMinor: 5,
	}
	
	tmpFile := "test_notebook.ipynb"
	data, _ := json.Marshal(nb)
	os.WriteFile(tmpFile, data, 0644)
	defer os.Remove(tmpFile)

	tool := &NotebookEditTool{}
	ctx := context.Background()

	t.Run("Insert a cell at index 1", func(t *testing.T) {
		args := map[string]interface{}{
			"file_path": tmpFile,
			"action":    "insert",
			"index":     1,
			"cell_type": "code",
			"source":    "x = 10",
		}
		
		_, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		// Verify
		updatedData, _ := os.ReadFile(tmpFile)
		var updatedNb JupyterNotebook
		json.Unmarshal(updatedData, &updatedNb)
		assert.Equal(t, 3, len(updatedNb.Cells))
		assert.Equal(t, "code", updatedNb.Cells[1].CellType)
	})

	t.Run("Remove a cell at index 0", func(t *testing.T) {
		args := map[string]interface{}{
			"file_path": tmpFile,
			"action":    "remove",
			"index":     0,
		}
		
		_, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		updatedData, _ := os.ReadFile(tmpFile)
		var updatedNb JupyterNotebook
		json.Unmarshal(updatedData, &updatedNb)
		assert.Equal(t, 2, len(updatedNb.Cells))
	})

    t.Run("Update a cell", func(t *testing.T) {
		args := map[string]interface{}{
			"file_path": tmpFile,
			"action":    "update",
			"index":     0,
            "cell_type": "markdown",
            "source": "# New Title",
		}
		
		_, err := tool.Execute(ctx, args)
		require.NoError(t, err)
		
		updatedData, _ := os.ReadFile(tmpFile)
		var updatedNb JupyterNotebook
		json.Unmarshal(updatedData, &updatedNb)
		assert.Equal(t, "# New Title", updatedNb.Cells[0].Source[0])
	})
}
