package cmd

import (
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync all tool configs from usher config",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, project, err := d.cfgLoader.LoadBoth()
			if err != nil {
				return err
			}
			return makeSyncFn()(global, project)
		},
	}
}
