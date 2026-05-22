package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vietnamesekid/usher/internal/actions"
)

func newMCPCmd() *cobra.Command {
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP servers",
	}
	mcpCmd.AddCommand(
		newMCPAddCmd(),
		newMCPRemoveCmd(),
		newMCPListCmd(),
	)
	return mcpCmd
}

func newMCPAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name>",
		Short: "Add an MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := actions.NewMCPActions(d.cfgLoader, d.cfgWriter, d.reg, d.kc, makeSyncFn(), d.out, d.prompt)
			return a.Add(args[0])
		},
	}
}

func newMCPRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an MCP server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := actions.NewMCPActions(d.cfgLoader, d.cfgWriter, d.reg, d.kc, makeSyncFn(), d.out, d.prompt)
			return a.Remove(args[0])
		},
	}
}

func newMCPListCmd() *cobra.Command {
	var available bool
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List configured MCP servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			a := actions.NewMCPActions(d.cfgLoader, d.cfgWriter, d.reg, d.kc, makeSyncFn(), d.out, d.prompt)
			if available {
				return a.ListAvailable()
			}
			return a.List()
		},
	}
	listCmd.Flags().BoolVar(&available, "available", false, "list all available MCPs in the registry")
	return listCmd
}
