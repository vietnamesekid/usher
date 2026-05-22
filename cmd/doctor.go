package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vietnamesekid/usher/internal/actions"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check the health of your usher configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			a := actions.NewDoctorActions(d.cfgLoader, d.kc, d.out)
			return a.Check()
		},
	}
}
