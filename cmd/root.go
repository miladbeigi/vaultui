package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/miladbeigi/vaultui/internal/app"
	"github.com/miladbeigi/vaultui/internal/config"
	"github.com/miladbeigi/vaultui/internal/vault"
)

var (
	cfgFile    string
	vaultAddr  string
	token      string
	namespace  string
	authMethod string
	username   string
	password   string
	roleID     string
	secretID   string
	authMount  string
)

var rootCmd = &cobra.Command{
	Use:   "vaultui",
	Short: "A k9s-inspired TUI for HashiCorp Vault",
	Long: `VaultUI is a keyboard-driven terminal UI for browsing, inspecting,
and managing HashiCorp Vault. Navigate secrets, policies, auth methods,
and leases — all without leaving the terminal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vc, err := vault.NewClient(vault.ClientConfig{
			Address:   viper.GetString("vault.address"),
			Token:     viper.GetString("vault.token"),
			Namespace: viper.GetString("vault.namespace"),
		})
		if err != nil {
			return fmt.Errorf("failed to create vault client: %w", err)
		}

		method := vault.AuthMethod(viper.GetString("auth.method"))
		if method != "" && method != vault.AuthToken {
			err := vc.Authenticate(vault.AuthConfig{
				Method:    method,
				MountPath: viper.GetString("auth.mount"),
				Username:  viper.GetString("auth.username"),
				Password:  viper.GetString("auth.password"),
				RoleID:    viper.GetString("auth.role-id"),
				SecretID:  viper.GetString("auth.secret-id"),
			})
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}
		}

		cfg, _ := config.Load(viper.GetString("config"))
		app.ApplyKeybindings(cfg.Settings.Keybindings)

		model := app.New(vc, cfg, viper.GetString("config"))
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run vaultui: %w", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.vaultui.yaml)")
	rootCmd.PersistentFlags().StringVar(&vaultAddr, "vault-addr", "", "Vault server address")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Vault authentication token")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "", "Vault namespace")
	rootCmd.PersistentFlags().StringVar(&authMethod, "auth-method", "token", "Auth method: token, userpass, approle")
	rootCmd.PersistentFlags().StringVar(&authMount, "auth-mount", "", "Custom mount path for auth method")
	rootCmd.PersistentFlags().StringVar(&username, "username", "", "Username for userpass auth")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "Password for userpass auth")
	rootCmd.PersistentFlags().StringVar(&roleID, "role-id", "", "Role ID for AppRole auth")
	rootCmd.PersistentFlags().StringVar(&secretID, "secret-id", "", "Secret ID for AppRole auth")

	_ = viper.BindPFlag("vault.address", rootCmd.PersistentFlags().Lookup("vault-addr"))
	_ = viper.BindPFlag("vault.token", rootCmd.PersistentFlags().Lookup("token"))
	_ = viper.BindPFlag("vault.namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	_ = viper.BindPFlag("auth.method", rootCmd.PersistentFlags().Lookup("auth-method"))
	_ = viper.BindPFlag("auth.mount", rootCmd.PersistentFlags().Lookup("auth-mount"))
	_ = viper.BindPFlag("auth.username", rootCmd.PersistentFlags().Lookup("username"))
	_ = viper.BindPFlag("auth.password", rootCmd.PersistentFlags().Lookup("password"))
	_ = viper.BindPFlag("auth.role-id", rootCmd.PersistentFlags().Lookup("role-id"))
	_ = viper.BindPFlag("auth.secret-id", rootCmd.PersistentFlags().Lookup("secret-id"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".vaultui")
		viper.SetConfigType("yaml")
	}

	// Environment variable overrides
	viper.SetEnvPrefix("")
	_ = viper.BindEnv("vault.address", "VAULT_ADDR")
	_ = viper.BindEnv("vault.token", "VAULT_TOKEN")
	_ = viper.BindEnv("vault.namespace", "VAULT_NAMESPACE")

	viper.AutomaticEnv()

	// Read config file (silently ignore if not found)
	_ = viper.ReadInConfig()
}
