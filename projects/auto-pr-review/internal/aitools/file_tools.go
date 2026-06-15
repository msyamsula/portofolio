package aitools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ReadFile struct {
	workspace string
}

func NewReadFile(workspace string) *ReadFile {
	return &ReadFile{workspace: workspace}
}

func (t *ReadFile) Name() string        { return "read_file" }
func (t *ReadFile) Description() string { return "Read the contents of a file in the workspace." }

func (t *ReadFile) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string", "description": "Relative path to the file in the workspace."},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFile) Execute(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	fullPath := filepath.Join(t.workspace, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}

type WriteFile struct {
	workspace string
}

func NewWriteFile(workspace string) *WriteFile {
	return &WriteFile{workspace: workspace}
}

func (t *WriteFile) Name() string        { return "write_file" }
func (t *WriteFile) Description() string { return "Write content to a file in the workspace." }

func (t *WriteFile) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path":    map[string]any{"type": "string", "description": "Relative path to the file in the workspace."},
			"content": map[string]any{"type": "string", "description": "Full file content to write."},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFile) Execute(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	content, _ := args["content"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	fullPath := filepath.Join(t.workspace, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(content), path), nil
}

type ListFiles struct {
	workspace string
}

func NewListFiles(workspace string) *ListFiles {
	return &ListFiles{workspace: workspace}
}

func (t *ListFiles) Name() string        { return "list_files" }
func (t *ListFiles) Description() string { return "List files in a directory within the workspace." }

func (t *ListFiles) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string", "description": "Directory to list, relative to workspace root. Empty for root."},
		},
	}
}

func (t *ListFiles) Execute(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	target := filepath.Join(t.workspace, path)

	var b strings.Builder
	filepath.Walk(target, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(p, ".git/") || strings.Contains(p, "node_modules/") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(t.workspace, p)
		if info.IsDir() {
			fmt.Fprintf(&b, "%s/\n", rel)
		} else {
			fmt.Fprintf(&b, "%s (%d bytes)\n", rel, info.Size())
		}
		return nil
	})

	if b.Len() == 0 {
		return "(empty)", nil
	}
	return b.String(), nil
}
