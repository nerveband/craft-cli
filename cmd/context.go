package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

type setting struct {
	Value  interface{} `json:"value"`
	Source string      `json:"source"`
}

func resolveREST() (string, string, map[string]setting, error) {
	values := map[string]setting{}
	url, key := apiURL, apiKey
	urlSource, keySource := "flag:api-url", "flag:api-key"
	if url == "" {
		url = os.Getenv("CRAFT_API_URL")
		urlSource = "env:CRAFT_API_URL"
	}
	if key == "" && apiKeyEnv != "" {
		key = os.Getenv(apiKeyEnv)
		keySource = "env:" + apiKeyEnv
		if key == "" {
			return "", "", nil, fmt.Errorf("--api-key-env %s is not set", apiKeyEnv)
		}
	}
	if key == "" {
		key = os.Getenv("CRAFT_API_KEY")
		keySource = "env:CRAFT_API_KEY"
	}
	selected := profileName
	profileSource := "unused:explicit-url"
	if url == "" {
		cfg, err := cfgManager.Load()
		if err != nil {
			return "", "", nil, err
		}
		profileSource = "flag:profile"
		if selected == "" {
			selected = cfg.ActiveProfile
			profileSource = "config:active_profile"
		}
		profile, ok := cfg.Profiles[selected]
		if !ok {
			return "", "", nil, newCLIError("CONFIG_ERROR", "no active profile; use profiles add-rest NAME --api-url URL")
		}
		if profile.TypeOrDefault() != "rest" {
			return "", "", nil, fmt.Errorf("profile %q is not rest", selected)
		}
		url, urlSource = profile.URL, "profile:"+selected
		if key == "" {
			key, keySource = profile.APIKey, "profile:"+selected
		}
	}
	if key == "" {
		keySource = "default:public"
	}
	values["profile"] = setting{selected, profileSource}
	values["api_url"] = setting{redactedURL(url), urlSource}
	values["api_key"] = setting{key != "", keySource}
	return url, key, values, nil
}

func effectiveConfig(cmd *cobra.Command) (map[string]interface{}, error) {
	cfg, err := cfgManager.Load()
	if err != nil {
		return nil, err
	}
	profiles, err := cfgManager.ListProfiles()
	if err != nil {
		return nil, err
	}
	values := map[string]setting{}
	_, _, rest, resolveErr := resolveREST()
	for k, v := range rest {
		values[k] = v
	}
	for name, value := range map[string]interface{}{"backend": backendName, "format": getOutputFormat(), "timeout": requestTimeout.String(), "data-source": dataSource, "deliver": deliverTarget} {
		source := "default"
		if cmd.Flags().Changed(name) {
			source = "flag:" + name
		} else if name == "format" && cfg.DefaultFormat != "" {
			source = cfgManager.DefaultFormatSource()
		}
		values[name] = setting{value, source}
	}
	mcp, mcpSource := mcpURL, "flag:mcp-url"
	if mcp == "" {
		mcp, mcpSource = os.Getenv("CRAFT_MCP_URL"), "env:CRAFT_MCP_URL"
	}
	selected := profileName
	if selected == "" {
		selected = cfg.ActiveProfile
	}
	if mcp == "" {
		mcp = cfg.Profiles[selected].MCPURL
		mcpSource = "profile:" + selected
	}
	if mcp == "" {
		mcpSource = "default:unset"
	}
	values["mcp_url"] = setting{redactedURL(mcp), mcpSource}
	configPath, credentialsPath := cfgManager.Paths()
	result := map[string]interface{}{"version": version, "settings": values, "profiles": safeProfiles(profiles), "config_path": configPath, "credentials_path": credentialsPath, "scope": "local"}
	if resolveErr != nil {
		result["rest_available"] = false
	} else {
		result["rest_available"] = true
	}
	return result, nil
}

var contextCmd = &cobra.Command{Use: "context", Short: "Show redacted runtime settings and available profiles", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
	value, err := effectiveConfig(cmd)
	if err != nil {
		return err
	}
	return outputJSON(value)
}}
var configInspectCmd = &cobra.Command{Use: "inspect", Short: "Show effective configuration and each setting's source", Args: cobra.NoArgs, RunE: contextCmd.RunE}

func init() {
	rootCmd.AddCommand(contextCmd)
	configCmd.AddCommand(configInspectCmd)
	commandEffects["context"] = CommandEffect{"read_only", "local", false}
	commandEffects["config inspect"] = CommandEffect{"read_only", "local", false}
}
