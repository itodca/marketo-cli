package cmd

import "github.com/spf13/cobra"

func newSetupCmd() *cobra.Command {
	return newAuthSetupCmd("setup", "Interactive alias for auth setup.", "setup is not implemented in the Go rewrite yet")
}
