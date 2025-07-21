package mcp

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"
)

// LoadMDCResources scans the given directory for .mdc files, parses the YAML
// front-matter, and returns a slice of ServerResource ready to be registered.
//
// Rules:
//   - Files must use the Obsidian-style front-matter delimiter `---`.
//   - Supported YAML keys: uri, name, mime. All optional.
//     – uri  (string)   : Resource URI to expose. Defaults to rules/<filename>.
//     – name (string)   : Human-readable name. Defaults to file base name.
//     – mime (string)   : MIME type, defaults to text/markdown.
//   - Content after the closing `---` is returned as the resource body.
func LoadMDCResources(dir string) ([]server.ServerResource, error) {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return nil, errors.New("rules directory not found")
	}

	var resources []server.ServerResource

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // abort walk
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".mdc") {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}
		raw := string(data)

		// Split frontmatter
		parts := strings.SplitN(raw, "---", 3)
		if len(parts) < 3 {
			return nil // invalid, skip
		}
		frontmatter := strings.TrimSpace(parts[1])
		body := strings.TrimSpace(parts[2])

		// Parse YAML.
		meta := struct {
			URI  string `yaml:"uri"`
			Name string `yaml:"name"`
			Mime string `yaml:"mime"`
		}{}
		_ = yaml.Unmarshal([]byte(frontmatter), &meta)

		base := strings.TrimSuffix(info.Name(), ".mdc")
		if meta.URI == "" {
			meta.URI = "obsidian-cli/rules/" + base
		}
		if meta.Name == "" {
			meta.Name = strings.ReplaceAll(base, "_", " ")
		}
		if meta.Mime == "" {
			meta.Mime = "text/markdown"
		}

		res := mcp.Resource{
			URI:      meta.URI,
			Name:     meta.Name,
			MIMEType: meta.Mime,
		}

		handler := func(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			return []mcp.ResourceContents{mcp.TextResourceContents{
				URI:      meta.URI,
				MIMEType: meta.Mime,
				Text:     body,
			}}, nil
		}

		resources = append(resources, server.ServerResource{Resource: res, Handler: handler})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resources, nil
}
