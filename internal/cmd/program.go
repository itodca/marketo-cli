package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newProgramCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	programCmd := &cobra.Command{
		Use:   "program",
		Short: "Program lookups.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	programCmd.AddCommand(newProgramListCmd(runtime, options))
	programCmd.AddCommand(newProgramGetCmd(runtime, options))

	return programCmd
}

func newProgramListCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		name  string
		limit int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List programs or fetch by exact name.",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			var result map[string]any
			if name != "" {
				result, err = apiClient.Get("/asset/v1/program/byName.json", map[string]any{"name": name})
			} else {
				result, err = apiClient.GetAllOffsetPages("/asset/v1/programs.json", nil, limit, 200)
			}
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&name, "name", "", "Lookup by exact program name.")
	flags.IntVar(&limit, "limit", 0, "Maximum number of records to return.")

	return cmd
}

func newProgramGetCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <program-id>",
		Short: "Get a program by id.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.Get(fmt.Sprintf("/asset/v1/program/%s.json", args[0]), nil)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}
}
