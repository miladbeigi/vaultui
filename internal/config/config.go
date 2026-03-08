package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Context represents a named Vault connection configuration.
type Context struct {
	Name      string `yaml:"name"`
	Address   string `yaml:"address"`
	Token     string `yaml:"token,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
	Auth      Auth   `yaml:"auth,omitempty"`
}

// Auth holds authentication-specific configuration for a context.
type Auth struct {
	Method    string `yaml:"method,omitempty"`
	MountPath string `yaml:"mount_path,omitempty"`
	Username  string `yaml:"username,omitempty"`
	Password  string `yaml:"password,omitempty"`
	RoleID    string `yaml:"role_id,omitempty"`
	SecretID  string `yaml:"secret_id,omitempty"`
}

// Config represents the full VaultUI configuration file.
type Config struct {
	CurrentContext string    `yaml:"current_context"`
	Contexts       []Context `yaml:"contexts"`
	Settings       Settings  `yaml:"settings,omitempty"`
}

// Settings holds global UI preferences.
type Settings struct {
	ClipboardTimeout int         `yaml:"clipboard_timeout,omitempty"`
	Keybindings      Keybindings `yaml:"keybindings,omitempty"`
}

// Keybindings allows users to override default key assignments.
// Values are comma-separated key names (e.g. "k,up").
type Keybindings struct {
	Up       string `yaml:"up,omitempty"`
	Down     string `yaml:"down,omitempty"`
	Top      string `yaml:"top,omitempty"`
	Bottom   string `yaml:"bottom,omitempty"`
	PageDown string `yaml:"page_down,omitempty"`
	PageUp   string `yaml:"page_up,omitempty"`
	Enter    string `yaml:"enter,omitempty"`
	Back     string `yaml:"back,omitempty"`
	Quit     string `yaml:"quit,omitempty"`
	Command  string `yaml:"command,omitempty"`
	Copy     string `yaml:"copy,omitempty"`
	CopyJSON string `yaml:"copy_json,omitempty"`
}

// DefaultConfigPath returns the default path for the config file.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".vaultui.yaml"
	}
	return filepath.Join(home, ".vaultui.yaml")
}

// Load reads the config file from the given path. Returns a zero Config
// if the file does not exist.
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %q: %w", path, err)
	}
	return &cfg, nil
}

// Save writes the config to the given path.
func Save(path string, cfg *Config) error {
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	return os.WriteFile(path, data, 0o600)
}

// GetContext returns the named context, or nil if not found.
func (c *Config) GetContext(name string) *Context {
	for i := range c.Contexts {
		if c.Contexts[i].Name == name {
			return &c.Contexts[i]
		}
	}
	return nil
}

// CurrentCtx returns the current context, or nil if not set.
func (c *Config) CurrentCtx() *Context {
	if c.CurrentContext == "" {
		return nil
	}
	return c.GetContext(c.CurrentContext)
}

// ContextNames returns the names of all configured contexts.
func (c *Config) ContextNames() []string {
	names := make([]string, len(c.Contexts))
	for i, ctx := range c.Contexts {
		names[i] = ctx.Name
	}
	return names
}
