package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var errUseBodyOrInput = errors.New("Use either --body or --input, not both")

func newAPICmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Raw API escape hatch.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	apiCmd.AddCommand(newAPIGetCmd(runtime, options))
	apiCmd.AddCommand(newAPIPostCmd(runtime, options))
	apiCmd.AddCommand(newAPIDeleteCmd(runtime, options))

	return apiCmd
}

func newAPIGetCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		queryValues []string
		fields      string
	)

	cmd := &cobra.Command{
		Use:   "get <path>",
		Short: "Send a raw GET request under /rest.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query, err := parseKVPairs(queryValues)
			if err != nil {
				return err
			}

			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.Get(args[0], query)
			if err != nil {
				return err
			}

			return writeResultFields(runtime, options, result, parseFields(fields))
		},
	}

	flags := cmd.Flags()
	flags.StringArrayVar(&queryValues, "query", nil, "Query parameter in key=value form.")
	flags.StringVar(&fields, "fields", "", "Comma-separated fields to return or display.")

	return cmd
}

func newAPIPostCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	return newAPIMutationCmd(runtime, options, "post", "Send a raw POST request under /rest.")
}

func newAPIDeleteCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	return newAPIMutationCmd(runtime, options, "delete", "Send a raw DELETE request under /rest.")
}

func newAPIMutationCmd(runtime *Runtime, options *RootOptions, use, short string) *cobra.Command {
	var (
		queryValues []string
		bodyValues  []string
		inputPath   string
		fields      string
	)

	cmd := &cobra.Command{
		Use:   use + " <path>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query, err := parseKVPairs(queryValues)
			if err != nil {
				return err
			}

			body, err := parseKVPairs(bodyValues)
			if err != nil {
				return err
			}

			inputBody, err := loadJSONInput(runtime, inputPath)
			if err != nil {
				return err
			}
			if body != nil && inputBody != nil {
				return errUseBodyOrInput
			}

			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			payload := body
			if inputBody != nil {
				payload = inputBody
			}

			var result map[string]any
			switch use {
			case "post":
				result, err = apiClient.Post(args[0], query, payload)
			case "delete":
				result, err = apiClient.Delete(args[0], query, payload)
			}
			if err != nil {
				return err
			}

			return writeResultFields(runtime, options, result, parseFields(fields))
		},
	}

	flags := cmd.Flags()
	flags.StringArrayVar(&queryValues, "query", nil, "Query parameter in key=value form.")
	flags.StringArrayVar(&bodyValues, "body", nil, "JSON body field in key=value form.")
	flags.StringVar(&inputPath, "input", "", "Path to a JSON object to send as the request body, or - for stdin.")
	flags.StringVar(&fields, "fields", "", "Comma-separated fields to return or display.")

	return cmd
}
