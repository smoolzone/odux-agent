package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagEventsLimit  int
	flagEventsServer int
	flagEventsAction string
)

// eventsCmd shows the unified audit + telemetry timeline.
var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Show the audit + telemetry timeline",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		out, err := c.ListEvents(at, flagEventsLimit, flagEventsServer, flagEventsAction)
		if err != nil {
			return err
		}
		if flagOutput == "json" {
			printJSON(out)
			return nil
		}
		if len(out) == 0 {
			fmt.Println("(no events)")
			return nil
		}
		for _, e := range out {
			icon := emoji(sword)
			if e.Kind == "audit" {
				icon = emoji(shield)
			}
			fmt.Printf("%s [%s] %s — %s\n", icon, e.CreatedAt, e.Action, e.Kind)
		}
		return nil
	},
}

func init() {
	eventsCmd.Flags().IntVar(&flagEventsLimit, "limit", 50, "max events to return")
	eventsCmd.Flags().IntVar(&flagEventsServer, "server", 0, "filter by chamber id")
	eventsCmd.Flags().StringVar(&flagEventsAction, "action", "", "filter by action")
	rootCmd.AddCommand(eventsCmd)
}
