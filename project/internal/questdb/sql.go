package questdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type execResponse struct {
	DDL    string `json:"ddl"`
	Error  string `json:"error"`
}

func (c *Client) ExecSQL(ctx context.Context, query string) error {
	endpoint := fmt.Sprintf(
		"http://%s:%d/exec?query=%s",
		c.cfg.Host,
		9000,
		url.QueryEscape(query),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("questdb http request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("questdb http exec: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("questdb read response: %w", err)
	}

	var result execResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("questdb decode response: %w", err)
	}

	if result.Error != "" {
		return fmt.Errorf("questdb sql error: %s", result.Error)
	}

	return nil
}

func (c *Client) MigrateDir(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read questdb migrations dir %s: %w", dir, err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		stmt := strings.TrimSpace(string(content))
		if stmt == "" {
			continue
		}

		if err := c.ExecSQL(ctx, stmt); err != nil {
			return fmt.Errorf("questdb migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}
