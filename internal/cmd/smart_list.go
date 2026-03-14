package cmd

import "github.com/spf13/cobra"

func newSmartListCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	smartListCmd := &cobra.Command{
		Use:   "smart-list",
		Short: "Smart list lookups.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	smartListCmd.AddCommand(newSmartListListCmd(runtime, options))
	smartListCmd.AddCommand(newSmartListGetCmd(runtime, options))

	return smartListCmd
}

func newSmartListListCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		name       string
		folderID   int
		folderType string
		limit      int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List smart lists or fetch by exact name.",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			var result map[string]any
			if name != "" {
				result, err = apiClient.Get("/asset/v1/smartList/byName.json", map[string]any{"name": name})
			} else {
				params := map[string]any{}
				folder, err := folderValue(folderID, folderType)
				if err != nil {
					return err
				}
				if folder != "" {
					params["folder"] = folder
				}
				result, err = apiClient.GetAllOffsetPages("/asset/v1/smartLists.json", paramsOrNil(params), limit, 200)
			}
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&name, "name", "", "Lookup by exact smart list name.")
	flags.IntVar(&folderID, "folder-id", 0, "Parent folder or program id.")
	flags.StringVar(&folderType, "folder-type", "", "Folder type: Folder or Program.")
	flags.IntVar(&limit, "limit", 0, "Maximum number of records to return.")

	return cmd
}

func newSmartListGetCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var includeRules bool

	cmd := &cobra.Command{
		Use:   "get <list-id>",
		Short: "Get a smart list by id.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			params := map[string]any{}
			if includeRules {
				params["includeRules"] = true
			}

			result, err := apiClient.Get("/asset/v1/smartList/"+args[0]+".json", paramsOrNil(params))
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	cmd.Flags().BoolVar(&includeRules, "include-rules", false, "Include smart list rules in the response.")
	return cmd
}
