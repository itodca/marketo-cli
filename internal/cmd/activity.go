package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newActivityCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	activityCmd := &cobra.Command{
		Use:   "activity",
		Short: "Activity lookups and lead changes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	activityCmd.AddCommand(newActivityTypesCmd(runtime, options))
	activityCmd.AddCommand(newActivityListCmd(runtime, options))
	activityCmd.AddCommand(newActivityChangesCmd(runtime, options))

	return activityCmd
}

func newActivityTypesCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "types",
		Short: "List Marketo activity types.",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			result, err := apiClient.Get("/v1/activities/types.json", nil)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}
}

func newActivityListCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		typeIDs []int
		since   int
		limit   int
	)

	cmd := &cobra.Command{
		Use:   "list <lead-id>",
		Short: "List activities for a lead.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			pagingToken, err := activityPagingToken(apiClient, since)
			if err != nil {
				return err
			}

			params := map[string]any{
				"leadIds":       args[0],
				"nextPageToken": pagingToken,
			}
			if len(typeIDs) > 0 {
				params["activityTypeIds"] = joinInts(typeIDs)
			}

			result, err := apiClient.GetAllPages("/v1/activities.json", params, limit, 0)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	flags := cmd.Flags()
	flags.IntSliceVar(&typeIDs, "type-id", nil, "Filter to activity type ids.")
	flags.IntVar(&since, "since", 30, "Days back from now.")
	flags.IntVar(&limit, "limit", 0, "Maximum number of records to return.")

	return cmd
}

func activityPagingToken(apiClient apiGetter, sinceDays int) (string, error) {
	result, err := apiClient.Get("/v1/activities/pagingtoken.json", map[string]any{
		"sinceDatetime": sinceDatetime(sinceDays),
	})
	if err != nil {
		return "", err
	}

	token, _ := result["nextPageToken"].(string)
	if token == "" {
		return "", fmt.Errorf("Marketo paging token response did not include nextPageToken")
	}
	return token, nil
}

func sinceDatetime(days int) string {
	return time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour).Format("2006-01-02T15:04:05Z")
}

func joinInts(values []int) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.Itoa(value))
	}
	return strings.Join(parts, ",")
}

type apiGetter interface {
	Get(path string, params map[string]any) (map[string]any, error)
	GetAllPages(path string, params map[string]any, limit int, batchSize int) (map[string]any, error)
}

func newActivityChangesCmd(runtime *Runtime, options *RootOptions) *cobra.Command {
	var (
		watch   []string
		leadIDs []int
		listID  int
		since   int
		limit   int
	)

	cmd := &cobra.Command{
		Use:   "changes",
		Short: "List lead change activities.",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := loadClient(runtime, options.Profile)
			if err != nil {
				return err
			}

			pagingToken, err := activityPagingToken(apiClient, since)
			if err != nil {
				return err
			}

			params := map[string]any{
				"fields":        strings.Join(watch, ","),
				"nextPageToken": pagingToken,
			}
			if len(leadIDs) > 0 {
				params["leadIds"] = joinInts(leadIDs)
			}
			if cmd.Flags().Changed("list-id") {
				params["listId"] = listID
			}

			result, err := apiClient.GetAllPages("/v1/activities/leadchanges.json", params, limit, 0)
			if err != nil {
				return err
			}

			return writeResult(runtime, options, result)
		},
	}

	flags := cmd.Flags()
	flags.StringArrayVar(&watch, "watch", nil, "Lead field name to watch. Repeat the option for multiple fields.")
	flags.IntSliceVar(&leadIDs, "lead-id", nil, "Filter to specific lead ids.")
	flags.IntVar(&listID, "list-id", 0, "Filter to a static list id.")
	flags.IntVar(&since, "since", 30, "Days back from now.")
	flags.IntVar(&limit, "limit", 0, "Maximum number of records to return.")
	_ = cmd.MarkFlagRequired("watch")

	return cmd
}
