package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newStaticListCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	staticListCmd := &cobra.Command{
		Use:   "static-list",
		Short: "Static list lookups.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	staticListCmd.AddCommand(newStaticListListCmd(runtime, options))
	staticListCmd.AddCommand(newStaticListGetCmd(runtime, options))
	staticListCmd.AddCommand(newStaticListMembersCmd(runtime, options))
	staticListCmd.AddCommand(newStaticListCheckCmd(runtime, options))

	return staticListCmd
}

func newStaticListListCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		name       string
		folderID   int
		folderType string
		limit      int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List static lists.",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			var result map[string]any
			if name != "" {
				result, err = apiClient.Get("/asset/v1/staticList/byName.json", map[string]any{"name": name})
			} else {
				params := map[string]any{}
				folder, err := folderValue(folderID, folderType)
				if err != nil {
					return err
				}
				if folder != "" {
					params["folder"] = folder
				}
				result, err = apiClient.GetAllOffsetPages("/asset/v1/staticLists.json", paramsOrNil(params), limit, 200)
			}
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&name, "name", "", "Lookup by exact static list name.")
	flags.IntVar(&folderID, "folder-id", 0, "Parent folder or program id.")
	flags.StringVar(&folderType, "folder-type", "", "Folder type: Folder or Program.")
	flags.IntVar(&limit, "limit", 0, "Maximum number of records to return.")

	return cmd
}

func newStaticListGetCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <list-id>",
		Short: "Get a static list by id.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
				if err != nil {
					return err
				}

				result, err := apiClient.Get("/asset/v1/staticList/"+args[0]+".json", nil)
				if err != nil {
					return err
				}

			return writeResult(runtime, options, result)
		},
	}
}

func newStaticListMembersCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		fields string
		limit  int
	)

	cmd := &cobra.Command{
		Use:   "members <list-id>",
		Short: "List static list members.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			params := map[string]any{}
			if fields != "" {
				params["fields"] = fields
			}

			result, err := apiClient.GetAllPages("/v1/lists/"+args[0]+"/leads.json", paramsOrNil(params), limit, 0)
			if err != nil {
				return err
			}

			return writeResultFields(runtime, options, result, parseFields(fields))
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&fields, "fields", "", "Comma-separated fields to return or display.")
	flags.IntVar(&limit, "limit", 0, "Maximum number of records to return.")

	return cmd
}

func newStaticListCheckCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var leadIDs []int

	cmd := &cobra.Command{
		Use:   "check <list-id>",
		Short: "Check whether leads belong to a static list.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(leadIDs) == 0 {
				return fmt.Errorf("At least one lead id is required")
			}
			if len(leadIDs) > 300 {
				return fmt.Errorf("A maximum of 300 leads is allowed per static list request")
			}

			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.Get("/v1/lists/"+args[0]+"/leads/ismember.json", map[string]any{"id": intsToStrings(leadIDs)})
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	cmd.Flags().IntSliceVar(&leadIDs, "lead", nil, "Lead id to check. Repeat the option for multiple leads.")
	return cmd
}

func paramsOrNil(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}
	return values
}

func intsToStrings(values []int) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, strconv.Itoa(value))
	}
	return result
}
