package sgbuddy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/spf13/cobra"

	"buddy/internal/apps/common"
	"buddy/internal/clients/datadog"
	"buddy/internal/di"
)

func NewDatadogCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dd",
		Short: "Datadog log utilities",
		Long:  "Interact with Datadog logs: search, aggregate, and submit log events.",
	}

	cmd.AddCommand(newDatadogSearchCmd(clients))
	cmd.AddCommand(newDatadogAggregateCmd(clients))
	cmd.AddCommand(newDatadogSubmitCmd(clients))

	return cmd
}

func newDatadogSearchCmd(clients *di.ClientSet) *cobra.Command {
	var params datadog.LogSearchParams
	var lastDays int

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search Datadog logs",
		Example: `  # Last 3 full days (exclusive of today)
  sgbuddy dd search "service:payment error" --last 3

  # Explicit ISO8601 window
  sgbuddy dd search --query "env:prod" --from "2025-11-01T00:00:00Z" --to "2025-11-02T00:00:00Z"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				params.Query = args[0]
			}
			if params.Query == "" {
				params.Query = "*"
			}
			if lastDays > 0 {
				if params.From != "" || params.To != "" {
					return fmt.Errorf("--last cannot be combined with --from/--to")
				}
				params.To = "now"
				params.From = fmt.Sprintf("now-%dd", lastDays)
			}

			res, err := clients.Datadog.SearchLogs(params)
			if err != nil {
				return err
			}

			output, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringVar(&params.Query, "query", "", "Log query")
	cmd.Flags().StringVar(&params.From, "from", "", "Start time, e.g. now-15m or 2025-11-01T00:00:00Z")
	cmd.Flags().StringVar(&params.To, "to", "", "End time, e.g. now or 2025-11-02T00:00:00Z")
	cmd.Flags().IntVar(&lastDays, "last", 0, "Look back this many whole days")
	cmd.Flags().StringVar(&params.Sort, "sort", "-timestamp", "Sort order: -timestamp or timestamp")
	cmd.Flags().IntVar(&params.Limit, "limit", 10, "Maximum number of logs to return")
	cmd.Flags().StringVar(&params.Cursor, "cursor", "", "Pagination cursor")
	cmd.Flags().StringSliceVar(&params.Indexes, "index", nil, "Restrict search to specific log indexes")

	return cmd
}

func newDatadogAggregateCmd(clients *di.ClientSet) *cobra.Command {
	var query string
	var from string
	var to string
	var indexes []string
	var aggregation string
	var interval string
	var computeType string

	cmd := &cobra.Command{
		Use:   "aggregate",
		Short: "Aggregate log metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			if query == "" {
				query = "*"
			}
			if from == "" {
				from = "now-15m"
			}
			if to == "" {
				to = "now"
			}
			nAgg := strings.ToLower(aggregation)
			agg, err := datadogV2.NewLogsAggregationFunctionFromValue(nAgg)
			if err != nil {
				return fmt.Errorf("invalid aggregation '%s': %w", aggregation, err)
			}
			compute := datadogV2.NewLogsCompute(*agg)
			nType := strings.ToLower(computeType)
			if nType != "" {
				ct, err := datadogV2.NewLogsComputeTypeFromValue(nType)
				if err != nil {
					return fmt.Errorf("invalid compute-type '%s': %w", computeType, err)
				}
				compute.SetType(*ct)
			}
			if interval != "" && (nType == "" || nType == string(datadogV2.LOGSCOMPUTETYPE_TIMESERIES)) {
				compute.SetInterval(interval)
			}

			filter := datadogV2.NewLogsQueryFilter()
			filter.SetQuery(query)
			filter.SetFrom(from)
			filter.SetTo(to)
			if len(indexes) > 0 {
				filter.SetIndexes(indexes)
			}

			request := datadogV2.NewLogsAggregateRequest()
			request.SetFilter(*filter)
			request.SetCompute([]datadogV2.LogsCompute{*compute})

			resp, _, err := clients.Datadog.AggregateLogs(*request)
			if err != nil {
				return err
			}

			output, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringVar(&query, "query", "*", "Log query for aggregation")
	cmd.Flags().StringVar(&from, "from", "now-15m", "Start time")
	cmd.Flags().StringVar(&to, "to", "now", "End time")
	cmd.Flags().StringSliceVar(&indexes, "index", nil, "Restrict aggregation to specific log indexes")
	cmd.Flags().StringVar(&aggregation, "aggregation", "count", "Aggregation function (e.g. count, avg)")
	cmd.Flags().StringVar(&interval, "interval", "5m", "Interval for timeseries buckets")
	cmd.Flags().StringVar(&computeType, "compute-type", "timeseries", "Compute type (timeseries or total)")

	return cmd
}

func newDatadogSubmitCmd(clients *di.ClientSet) *cobra.Command {
	var message string
	var serviceName string
	var hostname string

	cmd := &cobra.Command{
		Use:   "submit [message]",
		Short: "Submit logs to Datadog",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				message = args[0]
			}
			if message == "" {
				return fmt.Errorf("log message required")
			}

			item := datadogV2.NewHTTPLogItem(message)
			if serviceName != "" {
				item.SetService(serviceName)
			}
			if hostname != "" {
				item.SetHostname(hostname)
			}

			resp, _, err := clients.Datadog.SubmitLogs([]datadogV2.HTTPLogItem{*item}, nil)
			if err != nil {
				return err
			}
			fmt.Printf("Submit successful: %v\n", resp)
			return nil
		},
	}

	cmd.Flags().StringVar(&message, "message", "", "Log message")
	cmd.Flags().StringVar(&serviceName, "service", "", "Service name")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Origin hostname")

	return cmd
}
