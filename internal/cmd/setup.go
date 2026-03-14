package cmd

import "github.com/spf13/cobra"

func newSetupCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	return newAuthSetupCmd(runtime, options, "setup", "Interactive alias for auth setup.")
}
