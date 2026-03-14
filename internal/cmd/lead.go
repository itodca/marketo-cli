package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const defaultLeadFields = "id,email,firstName,lastName,company,unsubscribed,marketingSuspended,emailInvalid,sfdcLeadId,sfdcContactId,createdAt,updatedAt"

func newLeadCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	leadCmd := &cobra.Command{
		Use:   "lead",
		Short: "Lead lookups and memberships.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	leadCmd.AddCommand(newLeadGetCmd(runtime, options))
	leadCmd.AddCommand(newLeadListCmd(runtime, options))
	leadCmd.AddCommand(newLeadDescribeCmd(runtime, options))
	leadCmd.AddCommand(newLeadStaticListsCmd(runtime, options))
	leadCmd.AddCommand(newLeadProgramsCmd(runtime, options))
	leadCmd.AddCommand(newLeadSmartCampaignsCmd(runtime, options))

	return leadCmd
}

func newLeadGetCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var fields string

	cmd := &cobra.Command{
		Use:   "get <lead-id>",
		Short: "Get a lead by id.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.Get(fmt.Sprintf("/v1/lead/%s.json", args[0]), map[string]any{
				"fields": leadFields(fields),
			})
			if err != nil {
				return err
			}

			return writeResultFields(runtime, options, result, parseFields(fields))
		},
	}

	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated fields to return or display.")

	return cmd
}

func newLeadListCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		email       string
		leadIDs     string
		filterValue string
		fields      string
		limit       int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List leads by filter.",
		RunE: func(cmd *cobra.Command, args []string) error {
			filterType, filterValues, err := resolveLeadFilter(email, leadIDs, filterValue)
			if err != nil {
				return err
			}

			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.GetAllPages("/v1/leads.json", map[string]any{
				"filterType":   filterType,
				"filterValues": filterValues,
				"fields":       leadFields(fields),
			}, limit, 0)
			if err != nil {
				return err
			}

			return writeResultFields(runtime, options, result, parseFields(fields))
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&email, "email", "", "Filter by email address.")
	flags.StringVar(&leadIDs, "id", "", "Comma-separated Marketo lead ids.")
	flags.StringVar(&filterValue, "filter", "", "Custom filter as key=value.")
	flags.StringVar(&fields, "fields", "", "Comma-separated fields to return or display.")
	flags.IntVar(&limit, "limit", 0, "Maximum number of records to return.")

	return cmd
}

func leadFields(fields string) string {
	if strings.TrimSpace(fields) == "" {
		return defaultLeadFields
	}
	return fields
}

func resolveLeadFilter(email, leadIDs, filterValue string) (string, string, error) {
	switch {
	case email != "":
		return "email", email, nil
	case leadIDs != "":
		return "id", leadIDs, nil
	case filterValue != "":
		filterType, rawValue, ok := strings.Cut(filterValue, "=")
		if !ok {
			return "", "", fmt.Errorf("--filter must use key=value form")
		}
		filterType = strings.TrimSpace(filterType)
		rawValue = strings.TrimSpace(rawValue)
		if filterType == "" || rawValue == "" {
			return "", "", fmt.Errorf("--filter must use key=value form")
		}
		return filterType, rawValue, nil
	default:
		return "", "", fmt.Errorf("Provide one of --email, --id, or --filter")
	}
}

func newLeadDescribeCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var legacy bool

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe lead fields.",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			path := "/v1/leads/describe2.json"
			if legacy {
				path = "/v1/leads/describe.json"
			}

			result, err := apiClient.Get(path, nil)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	cmd.Flags().BoolVar(&legacy, "legacy", false, "Use the older describe endpoint instead of describe2.")
	return cmd
}

func newLeadStaticListsCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "static-lists <lead-id>",
		Short: "List static list memberships for a lead.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.GetAllPages("/v1/leads/"+args[0]+"/listMembership.json", nil, limit, 0)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of records to return.")
	return cmd
}

func newLeadProgramsCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		programIDs []int
		limit      int
	)

	cmd := &cobra.Command{
		Use:   "programs <lead-id>",
		Short: "List program memberships for a lead.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			params := map[string]any{}
			if len(programIDs) > 0 {
				params["filterType"] = "programId"
				params["filterValues"] = joinInts(programIDs)
			}

			result, err := apiClient.GetAllPages("/v1/leads/"+args[0]+"/programMembership.json", paramsOrNil(params), limit, 0)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	flags := cmd.Flags()
	flags.IntSliceVar(&programIDs, "program-id", nil, "Filter to specific program ids.")
	flags.IntVar(&limit, "limit", 0, "Maximum number of records to return.")

	return cmd
}

func newLeadSmartCampaignsCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "smart-campaigns <lead-id>",
		Short: "List smart campaign memberships for a lead.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.GetAllPages("/v1/leads/"+args[0]+"/smartCampaignMembership.json", nil, limit, 0)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of records to return.")
	return cmd
}
