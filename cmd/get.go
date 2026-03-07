package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/miladbeigi/vaultui/internal/vault"
)

var outputFormat string

var getCmd = &cobra.Command{
	Use:   "get [resource] [path]",
	Short: "Get Vault resources in headless mode (for scripting)",
	Long: `Retrieve Vault resources and output them as JSON.

Examples:
  vaultui get secret secret/apps/myapp/config
  vaultui get engines
  vaultui get policies
  vaultui get auth
  vaultui get health`,
	Args: cobra.MinimumNArgs(1),
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

		resource := args[0]
		switch resource {
		case "health":
			return getHealth(vc)
		case "engines":
			return getEngines(vc)
		case "auth":
			return getAuthMethods(vc)
		case "policies":
			return getPolicies(vc)
		case "secret":
			if len(args) < 2 {
				return fmt.Errorf("secret path required: vaultui get secret <mount/path>")
			}
			return getSecret(vc, args[1])
		case "policy":
			if len(args) < 2 {
				return fmt.Errorf("policy name required: vaultui get policy <name>")
			}
			return getPolicy(vc, args[1])
		default:
			return fmt.Errorf("unknown resource: %s (try: health, engines, auth, policies, secret, policy)", resource)
		}
	},
}

func init() {
	getCmd.Flags().StringVarP(&outputFormat, "output", "o", "json", "Output format (json)")
	rootCmd.AddCommand(getCmd)
}

func outputJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func getHealth(vc *vault.Client) error {
	health, err := vc.Health()
	if err != nil {
		return err
	}
	return outputJSON(health)
}

func getEngines(vc *vault.Client) error {
	engines, err := vc.ListSecretEngines()
	if err != nil {
		return err
	}
	return outputJSON(engines)
}

func getAuthMethods(vc *vault.Client) error {
	methods, err := vc.ListAuthMethods()
	if err != nil {
		return err
	}
	return outputJSON(methods)
}

func getPolicies(vc *vault.Client) error {
	policies, err := vc.ListPolicies()
	if err != nil {
		return err
	}
	return outputJSON(policies)
}

func getSecret(vc *vault.Client, fullPath string) error {
	engines, err := vc.ListSecretEngines()
	if err != nil {
		return fmt.Errorf("listing engines: %w", err)
	}

	var mount, subPath string
	for _, e := range engines {
		if len(fullPath) >= len(e.Path) && fullPath[:len(e.Path)] == e.Path {
			mount = e.Path
			subPath = fullPath[len(e.Path):]
			break
		}
	}
	if mount == "" {
		mount = fullPath
	}

	kvV2 := false
	for _, e := range engines {
		if e.Path == mount && e.Version == "v2" {
			kvV2 = true
			break
		}
	}

	data, err := vc.ReadSecret(mount, subPath, kvV2)
	if err != nil {
		return err
	}
	return outputJSON(data.Data)
}

func getPolicy(vc *vault.Client, name string) error {
	body, err := vc.GetPolicy(name)
	if err != nil {
		return err
	}
	return outputJSON(map[string]string{
		"name": name,
		"body": body,
	})
}
