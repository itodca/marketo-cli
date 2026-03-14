package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	lookPath           = exec.LookPath
	runExternalCommand = func(runtime *Runtime, args []string) error {
		dir, err := runtime.CurrentDir()
		if err != nil {
			return err
		}

		command := exec.Command(args[0], args[1:]...)
		command.Dir = dir
		command.Stdin = runtime.Stdin
		command.Stdout = runtime.Stdout
		command.Stderr = runtime.Stderr
		return command.Run()
	}
)

func newSkillCmd(runtime *Runtime) *cobra.Command {
	skillCmd := &cobra.Command{
		Use:   "skill",
		Short: "Agent skill installation.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	skillCmd.AddCommand(newSkillInstallCmd(runtime))
	skillCmd.AddCommand(newSkillUninstallCmd(runtime))

	return skillCmd
}

func newSkillInstallCmd(runtime *Runtime) *cobra.Command {
	var globalInstall bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the mrkto agent skill via npx skills.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSkillCommand(runtime, "add", globalInstall)
		},
	}

	cmd.Flags().BoolVar(&globalInstall, "global", false, "Install the skill for all projects.")
	return cmd
}

func newSkillUninstallCmd(runtime *Runtime) *cobra.Command {
	var globalInstall bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the mrkto agent skill via npx skills.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSkillCommand(runtime, "remove", globalInstall)
		},
	}

	cmd.Flags().BoolVar(&globalInstall, "global", false, "Uninstall the globally installed skill.")
	return cmd
}

func runSkillCommand(runtime *Runtime, action string, globalInstall bool) error {
	args := []string{"npx", "skills", action, "itodca/marketo-cli"}
	if globalInstall {
		args = append(args, "--global")
	}

	if _, err := lookPath(args[0]); err != nil {
		_, _ = fmt.Fprintln(runtime.Stderr, "npx not found. Install Node.js or run manually:")
		_, _ = fmt.Fprintf(runtime.Stderr, "  %s\n", strings.Join(args, " "))
		return &exitError{code: 1}
	}

	_, _ = fmt.Fprintf(runtime.Stdout, "Running: %s\n", strings.Join(args, " "))
	if err := runExternalCommand(runtime, args); err != nil {
		if exitCode := commandExitCode(err); exitCode != 0 {
			return &exitError{code: exitCode}
		}
		return err
	}

	return nil
}

func commandExitCode(err error) int {
	var exitCoder interface{ ExitCode() int }
	if err != nil && errors.As(err, &exitCoder) {
		return exitCoder.ExitCode()
	}
	return 1
}
