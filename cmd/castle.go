package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var flagSelector string

var castleCmd = &cobra.Command{
	Use:   "castle",
	Short: "🏰 List the chambers (servers) in your castle",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, at, err := session(cmd)
		if err != nil {
			return err
		}
		servers, err := c.Servers(at, flagSelector)
		if err != nil {
			return err
		}
		noun := "chamber"
		if len(servers) != 1 {
			noun = "chambers"
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintf(w, "%s Your castle — %d %s\n\n", emoji(castle), len(servers), noun)
		fmt.Fprintln(w, "ID\tNAME\tHOST\tROLE\tSTATUS")
		for _, s := range servers {
			fmt.Fprintf(w, "%d\t%s %s\t%s\t%s\t%s %s\n", s.ID, emoji(shield), s.Name, s.Host, s.Role, statusEmoji(s.Status), s.Status)
		}
		w.Flush()
		return nil
	},
}

func init() {
	castleCmd.Flags().StringVar(&flagSelector, "selector", "", "filter chambers by label selector (e.g. role=db,region=eu)")
	rootCmd.AddCommand(castleCmd)
}
