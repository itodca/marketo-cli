package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/itodca/marketo-cli/internal/profile"
)

func newAuthCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication and profile management.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	authCmd.AddCommand(newAuthListCmd(runtime, options))
	authCmd.AddCommand(newAuthCheckCmd(runtime, options))
	authCmd.AddCommand(newAuthSetupCmd("setup", "Write credentials for a profile.", "auth setup is not implemented in the Go rewrite yet"))

	return authCmd
}

func newAuthListCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured profiles.",
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

func newAuthSetupCmd(use, short, message string) *cobra.Command {
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
			_ = profileName
			_ = munchkinID
			_ = clientID
			_ = clientSecret
			_ = overwrite
			return errors.New(message)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&profileName, "profile", "", "Named profile to write.")
	flags.StringVar(&munchkinID, "munchkin-id", "", "Marketo munchkin id.")
	flags.StringVar(&clientID, "client-id", "", "LaunchPoint client id.")
	flags.StringVar(&clientSecret, "client-secret", "", "LaunchPoint client secret.")
	flags.BoolVar(&overwrite, "overwrite", false, "Overwrite the target profile if it already exists.")

	return cmd
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
