package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newCompanyCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	companyCmd := &cobra.Command{
		Use:   "company",
		Short: "Company lookups.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	companyCmd.AddCommand(newCompanyListCmd(runtime, options))
	companyCmd.AddCommand(newCompanyDescribeCmd(runtime, options))

	return companyCmd
}

func newCompanyListCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		name        string
		filterValue string
		fields      string
		limit       int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List companies by filter.",
		RunE: func(cmd *cobra.Command, args []string) error {
			filterType, filterValues, err := resolveCompanyFilter(name, filterValue)
			if err != nil {
				return err
			}

			params := map[string]any{
				"filterType":   filterType,
				"filterValues": filterValues,
			}
			if strings.TrimSpace(fields) != "" {
				params["fields"] = fields
			}

			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.GetAllPages("/v1/companies.json", params, limit, 0)
			if err != nil {
				return err
			}

			return writeResultFields(runtime, options, result, parseFields(fields))
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&name, "name", "", "Filter by company name.")
	flags.StringVar(&filterValue, "filter", "", "Custom filter as key=value.")
	flags.StringVar(&fields, "fields", "", "Comma-separated fields to return or display.")
	flags.IntVar(&limit, "limit", 0, "Maximum number of records to return.")

	return cmd
}

func newCompanyDescribeCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "describe",
		Short: "Describe company fields.",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.Get("/v1/companies/describe.json", nil)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}
}

func resolveCompanyFilter(name, filterValue string) (string, string, error) {
	switch {
	case name != "":
		return "company", name, nil
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
		return "", "", fmt.Errorf("Provide one of --name or --filter")
	}
}
