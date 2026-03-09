// Package models — dynamic model discovery.
//
// Sources:
//
//	Ollama Registry — https://registry.ollama.ai/v2/library/{model}/tags/list
//	Ollama Library  — https://ollama.com/search
//	GitHub API      — https://api.github.com/search/repositories
package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	ollamaSearchURL   = "https://ollama.com/search"
	ollamaRegistryURL = "https://registry.ollama.ai/v2/library"
	githubSearchURL   = "https://api.github.com/search/repositories"
	fetchTimeout      = 10 * time.Second
	maxBodyBytes      = 2 << 20 // 2 MB
)

// DiscoveredModel represents a model found from a remote source.
type DiscoveredModel struct {
	Name        string   `json:"name"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`
	Stars       int      `json:"stars,omitempty"`
	Source      string   `json:"source"` // "ollama", "ollama-registry", "github"
	SourceURL   string   `json:"source_url,omitempty"`
}

// FetchOllamaModels discovers models from the Ollama library.
// It tries scraping the search page first, then falls back to querying
// the registry for known model families from the static catalog.
func FetchOllamaModels(ctx context.Context) ([]DiscoveredModel, error) {
	client := &http.Client{Timeout: fetchTimeout}

	// Strategy 1: scrape the Ollama search page for model names.
	if models, err := fetchOllamaSearchPage(ctx, client); err == nil && len(models) > 0 {
		return models, nil
	}

	// Strategy 2: query registry tag lists for every family in static catalog.
	return fetchKnownFamilies(ctx, client)
}

// fetchOllamaSearchPage fetches https://ollama.com/search and extracts model names.
func fetchOllamaSearchPage(ctx context.Context, client *http.Client) ([]DiscoveredModel, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ollamaSearchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "lokai/1.0 (https://github.com/romeo-mz/lokai)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama search: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, err
	}

	return parseOllamaHTML(string(body)), nil
}

// parseOllamaHTML extracts model names from the Ollama search page HTML.
func parseOllamaHTML(html string) []DiscoveredModel {
	// Match href patterns like /library/<model> or /<model> (Ollama URL patterns).
	re := regexp.MustCompile(`href="/(library/)?([a-z][a-z0-9._-]{1,39})"`)
	matches := re.FindAllStringSubmatch(html, -1)

	skip := map[string]bool{
		"blog": true, "download": true, "search": true, "signup": true,
		"login": true, "about": true, "privacy": true, "terms": true,
		"docs": true, "api": true, "discord": true, "github": true,
	}

	seen := make(map[string]bool)
	var out []DiscoveredModel
	for _, m := range matches {
		name := m[2]
		if skip[name] || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, DiscoveredModel{
			Name:      name,
			Source:    "ollama",
			SourceURL: "https://ollama.com/library/" + name,
		})
	}
	return out
}

// fetchKnownFamilies queries the Ollama registry for tags of each family
// already present in the static catalog.
func fetchKnownFamilies(ctx context.Context, client *http.Client) ([]DiscoveredModel, error) {
	families := knownModelFamilies()
	var out []DiscoveredModel

	for _, fam := range families {
		tags, err := fetchRegistryTags(ctx, client, fam)
		if err != nil {
			continue // best-effort
		}
		out = append(out, DiscoveredModel{
			Name:      fam,
			Tags:      tags,
			Source:    "ollama-registry",
			SourceURL: "https://ollama.com/library/" + fam,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no models found via registry fallback")
	}
	return out, nil
}

// fetchRegistryTags returns the available tags for a model from the OCI registry.
func fetchRegistryTags(ctx context.Context, client *http.Client, model string) ([]string, error) {
	url := fmt.Sprintf("%s/%s/tags/list", ollamaRegistryURL, model)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry HTTP %d for %s", resp.StatusCode, model)
	}

	var result struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBodyBytes)).Decode(&result); err != nil {
		return nil, err
	}
	return result.Tags, nil
}

// FetchGitHubModels searches GitHub for Ollama-compatible model repositories.
// Source: https://api.github.com/search/repositories
func FetchGitHubModels(ctx context.Context) ([]DiscoveredModel, error) {
	client := &http.Client{Timeout: fetchTimeout}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		githubSearchURL+"?q=ollama+model+in:topics&sort=stars&per_page=30", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "lokai/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github search: HTTP %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Stars       int    `json:"stargazers_count"`
			HTMLURL     string `json:"html_url"`
		} `json:"items"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBodyBytes)).Decode(&result); err != nil {
		return nil, err
	}

	var out []DiscoveredModel
	for _, item := range result.Items {
		out = append(out, DiscoveredModel{
			Name:        item.Name,
			Description: item.Description,
			Stars:       item.Stars,
			Source:      "github",
			SourceURL:   item.HTMLURL,
		})
	}
	return out, nil
}

// knownModelFamilies extracts unique model families from the static catalog.
func knownModelFamilies() []string {
	seen := make(map[string]bool)
	var families []string
	for _, m := range Catalog {
		fam := strings.Split(m.OllamaTag, ":")[0]
		if !seen[fam] {
			seen[fam] = true
			families = append(families, fam)
		}
	}
	return families
}
