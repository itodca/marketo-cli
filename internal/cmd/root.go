package cmd

import (
	"github.com/spf13/cobra"

	"github.com/itodca/marketo-cli/internal/version"
)

func NewRootCmd(runtime *Runtime) *cobra.Command {
	options := &RootOptions{}

	rootCmd := &cobra.Command{
		Use:           "mrkto",
		Short:         "Marketo REST API CLI.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.SetOut(runtime.Stdout)
	rootCmd.SetErr(runtime.Stderr)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	flags := rootCmd.PersistentFlags()
	flags.StringVar(&options.Profile, "profile", "", "Named profile to use.")
	flags.BoolVar(&options.JSON, "json", false, "Pretty JSON output (default).")
	flags.BoolVar(&options.Compact, "compact", false, "One JSON object per line.")
	flags.BoolVar(&options.Raw, "raw", false, "Single-line JSON output for the full returned payload.")

	rootCmd.AddCommand(newAuthCmd(runtime, options))
	rootCmd.AddCommand(newAPICmd(runtime, options))
	rootCmd.AddCommand(newSetupCmd())
	rootCmd.AddCommand(newVersionCmd(runtime))

	return rootCmd
}

func Execute() int {
	runtime := NewRuntime()
	if err := NewRootCmd(runtime).Execute(); err != nil {
		writeError(runtime, err)
		return 1
	}
	return 0
}

func newVersionCmd(runtime *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeResult(runtime, &RootOptions{JSON: true}, map[string]any{
				"version": version.Version,
				"commit":  version.Commit,
				"date":    version.Date,
			})
		},
	}
}
