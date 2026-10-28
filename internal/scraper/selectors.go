package scraper

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ListSelectors defines CSS selectors for extracting item lists.
type ListSelectors struct {
	Item     string `yaml:"item"`
	Title    string `yaml:"title"`
	Link     string `yaml:"link"`
	Date     string `yaml:"date"`
	NextPage string `yaml:"next_page"`
}

// DetailSelectors defines CSS selectors for extracting detail fields.
type DetailSelectors struct {
	Abstract    string `yaml:"abstract"`
	ReleaseDate string `yaml:"release_date"`
	FileSize    string `yaml:"file_size"`
	FileLink    string `yaml:"file_link"`
	ImageLink   string `yaml:"image_link"`
}

// KindSelector defines selectors for a specific Kind.
type KindSelector struct {
	ListURL         string          `yaml:"list_url"`
	List            ListSelectors   `yaml:"list"`
	Detail          DetailSelectors `yaml:"detail"`
	RequiredMarkers []string        `yaml:"required_markers"`
}

// SelectorsConfig maps Kind names to their respective KindSelector.
type SelectorsConfig struct {
	Version int                     `yaml:"version"`
	Kinds   map[string]KindSelector `yaml:",inline"`
}

// LoadSelectors reads and parses the YAML selector configuration (§8.3).
func LoadSelectors(path string, enabledKinds []Kind) (*SelectorsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("selectors: failed to read file %q: %w", path, err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("selectors: invalid yaml syntax: %w", err)
	}

	cfg := &SelectorsConfig{
		Kinds: make(map[string]KindSelector),
	}

	for k, v := range raw {
		if k == "version" {
			if ver, ok := v.(int); ok {
				cfg.Version = ver
			}
			continue
		}
		// Marshal kind back to YAML and unmarshal into KindSelector
		kindBytes, err := yaml.Marshal(v)
		if err != nil {
			continue
		}
		var ks KindSelector
		if err := yaml.Unmarshal(kindBytes, &ks); err == nil {
			cfg.Kinds[k] = ks
		}
	}

	// Validate that active enabled kinds do not have "TBD" (§8.3)
	for _, ek := range enabledKinds {
		kStr := string(ek)
		ks, ok := cfg.Kinds[kStr]
		if !ok {
			return nil, fmt.Errorf("selectors: configuration for enabled kind %q not found in %s", kStr, path)
		}
		if hasTBD(ks.ListURL) || hasTBD(ks.List.Item) || hasTBD(ks.List.Title) || hasTBD(ks.List.Link) {
			return nil, fmt.Errorf("selectors: kind %q contains unverified TBD selectors in %s", kStr, path)
		}
	}

	return cfg, nil
}

// ForKind returns the KindSelector for a kind, or an error if missing.
func (c *SelectorsConfig) ForKind(k Kind) (KindSelector, error) {
	if ks, ok := c.Kinds[string(k)]; ok {
		return ks, nil
	}
	return KindSelector{}, fmt.Errorf("selectors: no selectors configured for kind %s", k)
}

func hasTBD(s string) bool {
	return strings.Contains(strings.ToUpper(strings.TrimSpace(s)), "TBD")
}
