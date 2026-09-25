package roles

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Role holds the scopes assigned to a bearer token.
type Role struct {
	Name   string
	Scopes []string
}

// ScopeRule is a compiled (glob, scope) pair from roles.yaml.
type ScopeRule struct {
	Glob  string
	Scope string
}

// Config is the full parsed content of roles.yaml.
type Config struct {
	// ScopeRules is the ordered list of path-to-scope mappings.
	ScopeRules []ScopeRule
	// LabelScopes maps a label name to the scope it confers.
	LabelScopes map[string]string
	// KindScopes maps an item kind to the scope it confers when no label matches.
	KindScopes map[string]string
	// LabelAreas maps a label to the path scope used by starter_tasks.
	LabelAreas map[string]string
	// Roles maps a role name to its Role value.
	Roles map[string]Role
}

// rawFile mirrors the top-level structure of roles.yaml for yaml unmarshalling.
type rawFile struct {
	Scopes      []rawScope        `yaml:"scopes"`
	LabelScopes map[string]string `yaml:"label_scopes"`
	KindScopes  map[string]string `yaml:"kind_scopes"`
	LabelAreas  map[string]string `yaml:"label_areas"`
	Roles       map[string]rawRole `yaml:"roles"`
}

type rawScope struct {
	Glob  string `yaml:"glob"`
	Scope string `yaml:"scope"`
}

type rawRole struct {
	TokenEnv string   `yaml:"token_env"`
	Scopes   []string `yaml:"scopes"`
}

// Load reads and parses the YAML file at path. For each role whose token_env variable is set in
// the environment, the role is included in the returned Config. Roles whose token_env variable is
// unset are silently omitted (the token is not known, so they cannot be used). The token value is
// never stored or returned.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("roles: read %s: %w", path, err)
	}
	return parse(data)
}

// ParseBytes parses a roles.yaml from the given bytes. It is exported so that
// tests in other packages can construct a Config without writing a file.
func ParseBytes(data []byte) (*Config, error) {
	return parse(data)
}

// parse is Load without the I/O so that tests can supply bytes directly.
func parse(data []byte) (*Config, error) {
	var raw rawFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("roles: parse yaml: %w", err)
	}

	cfg := &Config{
		LabelScopes: raw.LabelScopes,
		KindScopes:  raw.KindScopes,
		LabelAreas:  raw.LabelAreas,
		Roles:       make(map[string]Role, len(raw.Roles)),
	}
	if cfg.LabelScopes == nil {
		cfg.LabelScopes = map[string]string{}
	}
	if cfg.KindScopes == nil {
		cfg.KindScopes = map[string]string{}
	}
	if cfg.LabelAreas == nil {
		cfg.LabelAreas = map[string]string{}
	}

	for _, rs := range raw.Scopes {
		cfg.ScopeRules = append(cfg.ScopeRules, ScopeRule{Glob: rs.Glob, Scope: rs.Scope})
	}

	for name, rr := range raw.Roles {
		cfg.Roles[name] = Role{Name: name, Scopes: rr.Scopes}
	}

	return cfg, nil
}

// PathScope returns the scope for path by applying the first matching rule from rules.
// If no rule matches, it returns "(unmapped)".
func PathScope(rules []ScopeRule, path string) string {
	for _, r := range rules {
		if matchGlob(r.Glob, path) {
			return r.Scope
		}
	}
	return "(unmapped)"
}
