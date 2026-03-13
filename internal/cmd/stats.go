package cmd

import "github.com/spf13/cobra"

func newStatsCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Usage and error stats.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	statsCmd.AddCommand(newStatsUsageCmd(runtime, options))
	statsCmd.AddCommand(newStatsErrorsCmd(runtime, options))

	return statsCmd
}

func newStatsUsageCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var weekly bool

	cmd := &cobra.Command{
		Use:   "usage",
		Short: "Show Marketo API usage stats.",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			path := "/v1/stats/usage.json"
			if weekly {
				path = "/v1/stats/usage/last7days.json"
			}

			result, err := apiClient.Get(path, nil)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	cmd.Flags().BoolVar(&weekly, "weekly", false, "Return the last 7 days instead of the current period.")
	return cmd
}

func newStatsErrorsCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var weekly bool

	cmd := &cobra.Command{
		Use:   "errors",
		Short: "Show Marketo API error stats.",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			path := "/v1/stats/errors.json"
			if weekly {
				path = "/v1/stats/errors/last7days.json"
			}

			result, err := apiClient.Get(path, nil)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	cmd.Flags().BoolVar(&weekly, "weekly", false, "Return the last 7 days instead of the current period.")
	return cmd
}
