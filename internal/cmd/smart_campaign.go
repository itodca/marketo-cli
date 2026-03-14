package cmd

import (
	"errors"
	"fmt"

	"github.com/itodca/marketo-cli/internal/client"

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
	smartCampaignCmd.AddCommand(newSmartCampaignScheduleCmd(runtime, options))
	smartCampaignCmd.AddCommand(newSmartCampaignTriggerCmd(runtime, options))

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

func newSmartCampaignScheduleCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		runAt   string
		execute bool
	)

	cmd := &cobra.Command{
		Use:   "schedule <campaign-id>",
		Short: "Schedule a smart campaign.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			campaignID, err := parseIntArg("campaign id", args[0])
			if err != nil {
				return err
			}

			request := map[string]any{"input": map[string]any{}}
			if runAt != "" {
				request["input"].(map[string]any)["runAt"] = runAt
			}

			if !execute {
				return writeResult(runtime, options, map[string]any{
					"dry_run":     true,
					"resource":    "smart-campaign",
					"action":      "schedule",
					"campaign_id": campaignID,
					"request":     request,
				})
			}

			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.Post("/v1/campaigns/"+args[0]+"/schedule.json", nil, request)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&runAt, "run-at", "", "Schedule execution time in Marketo-compatible format.")
	flags.BoolVar(&execute, "execute", false, "Send the request instead of returning a dry-run payload.")

	return cmd
}

func newSmartCampaignTriggerCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		leadIDs []int
		execute bool
	)

	cmd := &cobra.Command{
		Use:   "trigger <campaign-id>",
		Short: "Trigger a smart campaign for one or more leads.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			campaignID, err := parseIntArg("campaign id", args[0])
			if err != nil {
				return err
			}

			if err := validateLeadIDs(leadIDs, 100, "trigger request"); err != nil {
				return err
			}

			request := map[string]any{
				"input": map[string]any{
					"leads": leadInputs(leadIDs),
				},
			}

			if !execute {
				return writeResult(runtime, options, map[string]any{
					"dry_run":     true,
					"resource":    "smart-campaign",
					"action":      "trigger",
					"campaign_id": campaignID,
					"request":     request,
				})
			}

			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.Post("/v1/campaigns/"+args[0]+"/trigger.json", nil, request)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	flags := cmd.Flags()
	flags.IntSliceVar(&leadIDs, "lead", nil, "Lead id to trigger. Repeat the option for multiple leads.")
	flags.BoolVar(&execute, "execute", false, "Send the request instead of returning a dry-run payload.")

	return cmd
}

func validateLeadIDs(leadIDs []int, max int, actionLabel string) error {
	if len(leadIDs) == 0 {
		return fmt.Errorf("At least one lead id is required")
	}
	if len(leadIDs) > max {
		return &client.APIError{
			Code:    "invalid_input",
			Message: fmt.Sprintf("A maximum of %d leads is allowed per %s", max, actionLabel),
		}
	}
	return nil
}

func leadInputs(leadIDs []int) []map[string]int {
	inputs := make([]map[string]int, 0, len(leadIDs))
	for _, leadID := range leadIDs {
		inputs = append(inputs, map[string]int{"id": leadID})
	}
	return inputs
}
