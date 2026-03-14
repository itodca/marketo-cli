package cmd

import (
	"github.com/spf13/cobra"

	"github.com/itodca/marketo-cli/internal/config"
	"github.com/itodca/marketo-cli/internal/profile"
)

func newAuthCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication and saved Marketo connections.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	authCmd.AddCommand(newAuthListCmd(runtime, options))
	authCmd.AddCommand(newAuthCheckCmd(runtime, options))
	authCmd.AddCommand(newAuthSetupCmd(runtime, options, "setup", "Save credentials for a Marketo instance or environment."))

	return authCmd
}

func newAuthListCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List saved Marketo connections.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profiles, err := profile.ListProfiles()
			if err != nil {
				return err
			}

			result := make([]map[string]any, 0, len(profiles))
			for _, profileName := range profiles {
				result = append(result, map[string]any{"profile": profileName})
			}

			return writeResult(runtime, options, map[string]any{
				"success": true,
				"result":  result,
			})
		},
	}
}

func newAuthSetupCmd(runtime *Runtime, options *RootOptions, use, short string) *cobra.Command {
	var (
		profileName  string
		munchkinID   string
		clientID     string
		clientSecret string
		overwrite    bool
	)

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedMunchkinID, err := promptIfMissing(runtime, munchkinID, "Munchkin ID")
			if err != nil {
				return err
			}
			resolvedClientID, err := promptIfMissing(runtime, clientID, "Client ID")
			if err != nil {
				return err
			}
			resolvedClientSecret, err := promptIfMissing(runtime, clientSecret, "Client Secret")
			if err != nil {
				return err
			}

			configPath, err := config.Write(profileName, resolvedMunchkinID, resolvedClientID, resolvedClientSecret, overwrite)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, map[string]any{
				"status":  "saved",
				"path":    configPath,
				"profile": profileNameOrDefault(profileName),
			})
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&profileName, "profile", "", "Profile name to save, such as sandbox, production, or another instance.")
	flags.StringVar(&munchkinID, "munchkin-id", "", "Marketo munchkin id.")
	flags.StringVar(&clientID, "client-id", "", "LaunchPoint client id.")
	flags.StringVar(&clientSecret, "client-secret", "", "LaunchPoint client secret.")
	flags.BoolVar(&overwrite, "overwrite", false, "Overwrite the target profile if it already exists.")

	return cmd
}

func profileNameOrDefault(profileName string) string {
	if profileName == "" {
		return "default"
	}
	return profileName
}

func newAuthCheckCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Validate the current profile by making an authenticated API request.",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.Get("/v1/activities/types.json", nil)
			if err != nil {
				return err
			}

			activityTypesAvailable := 0
			if records, ok := result["result"].([]any); ok {
				activityTypesAvailable = len(records)
			}

			return writeResult(runtime, options, map[string]any{
				"success": true,
				"result": []map[string]any{
					{
						"status":                   "ok",
						"profile":                  apiClient.Config.Profile,
						"munchkin_id":              apiClient.Config.MunchkinID,
						"rest_url":                 apiClient.Config.RestURL,
						"activity_types_available": activityTypesAvailable,
					},
				},
			})
		},
	}
}
