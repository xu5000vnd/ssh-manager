package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"ssh-manager/internal/config"
)

var (
	listJSON  bool
	listGroup string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved SSH connections",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _, _, err := loadConfigWithPassword()
		if err != nil {
			return err
		}

		conns := make([]config.Connection, 0, len(cfg.Connections))
		for _, c := range cfg.Connections {
			if listGroup == "" || strings.EqualFold(c.Group, listGroup) {
				conns = append(conns, c)
			}
		}

		sort.Slice(conns, func(i, j int) bool { return strings.ToLower(conns[i].Name) < strings.ToLower(conns[j].Name) })

		if listJSON {
			b, err := json.MarshalIndent(conns, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tGROUP\tHOST\tUSER\tPORT")
		for _, c := range conns {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", c.Name, c.Group, c.Host, c.Username, c.Port)
		}
		return w.Flush()
	},
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "output as JSON")
	listCmd.Flags().StringVar(&listGroup, "group", "", "filter by group")
}
