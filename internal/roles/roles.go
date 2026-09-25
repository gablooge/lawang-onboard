package roles

import (
	"crypto/subtle"
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
	// tokenEnvs maps role name to the environment variable that holds its token.
	// This is stored so that RoleForToken can resolve tokens without re-parsing.
	tokenEnvs map[string]string
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

// RoleForToken looks up the role for the given bearer token by reading each
// role's token_env environment variable at call time. It returns the matching
// Role and true, or the zero Role and false if the token is unknown, empty,
// or matches more than one role (fail closed on ambiguity).
// The token value is never logged or returned in error messages.
func (cfg *Config) RoleForToken(token string) (Role, bool) {
	if token == "" {
		return Role{}, false
	}
	var matched Role
	count := 0
	for name, envVar := range cfg.tokenEnvs {
		if envVar == "" {
			continue
		}
		val := os.Getenv(envVar)
		if val == "" {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(val), []byte(token)) == 1 {
			r, ok := cfg.Roles[name]
			if !ok {
				continue
			}
			matched = r
			count++
		}
	}
	if count == 1 {
		return matched, true
	}
	return Role{}, false
}

// TokenEnvVar returns the environment variable name that holds the token for
// the named role, or the empty string if the role is not defined.
func (cfg *Config) TokenEnvVar(roleName string) string {
	return cfg.tokenEnvs[roleName]
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

	cfg.tokenEnvs = make(map[string]string, len(raw.Roles))
	// tokenValues tracks env var name -> role name, used to detect duplicate token vars.
	tokenValues := make(map[string]string, len(raw.Roles))
	for name, rr := range raw.Roles {
		cfg.Roles[name] = Role{Name: name, Scopes: rr.Scopes}
		cfg.tokenEnvs[name] = rr.TokenEnv
		if rr.TokenEnv != "" {
			if prev, dup := tokenValues[rr.TokenEnv]; dup {
				return nil, fmt.Errorf("roles: %s and %s share the same token_env variable %s", prev, name, rr.TokenEnv)
			}
			tokenValues[rr.TokenEnv] = name
		}
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
