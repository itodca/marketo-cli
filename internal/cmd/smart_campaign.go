package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var errActiveAndAll = errors.New("Choose only one of --active or --all")

func newSmartCampaignCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	smartCampaignCmd := &cobra.Command{
		Use:   "smart-campaign",
		Short: "Smart campaign lookups.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	smartCampaignCmd.AddCommand(newSmartCampaignListCmd(runtime, options))
	smartCampaignCmd.AddCommand(newSmartCampaignGetCmd(runtime, options))

	return smartCampaignCmd
}

func newSmartCampaignListCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		name       string
		folderID   int
		folderType string
		active     bool
		all        bool
		limit      int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List smart campaigns or fetch by exact name.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if active && all {
				return errActiveAndAll
			}

			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			var result map[string]any
			if name != "" {
				result, err = apiClient.Get("/asset/v1/smartCampaign/byName.json", map[string]any{"name": name})
			} else {
				params := map[string]any{}
				folder, err := folderValue(folderID, folderType)
				if err != nil {
					return err
				}
				if folder != "" {
					params["folder"] = folder
				}
				if active {
					params["isActive"] = true
				}
				result, err = apiClient.GetAllOffsetPages("/asset/v1/smartCampaigns.json", paramsOrNil(params), limit, 200)
			}
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&name, "name", "", "Lookup by exact smart campaign name.")
	flags.IntVar(&folderID, "folder-id", 0, "Parent folder or program id.")
	flags.StringVar(&folderType, "folder-type", "", "Folder type: Folder or Program.")
	flags.BoolVar(&active, "active", false, "Only active smart campaigns.")
	flags.BoolVar(&all, "all", false, "Request all smart campaigns, including inactive ones.")
	flags.IntVar(&limit, "limit", 0, "Maximum number of records to return.")

	return cmd
}

func newSmartCampaignGetCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <campaign-id>",
		Short: "Get a smart campaign by id.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.Get("/asset/v1/smartCampaign/"+args[0]+".json", nil)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}
}
