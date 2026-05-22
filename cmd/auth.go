package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vietnamesekid/usher/internal/actions"
)

func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage API keys for AI providers",
	}
	authCmd.AddCommand(newAuthSetupCmd(), newAuthRevokeCmd())
	return authCmd
}

func newAuthSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Interactive wizard to configure API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			a := actions.NewAuthActions(d.cfgLoader, d.cfgWriter, d.kc, d.out, d.prompt)
			return a.Setup()
		},
	}
}

func newAuthRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <provider>",
		Short: "Revoke credentials for a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := actions.NewAuthActions(d.cfgLoader, d.cfgWriter, d.kc, d.out, d.prompt)
			return a.Revoke(args[0])
		},
	}
}
