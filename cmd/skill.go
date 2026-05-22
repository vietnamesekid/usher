package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vietnamesekid/usher/internal/actions"
)

func newSkillCmd() *cobra.Command {
	skillCmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage AI skills via skills.sh",
	}
	skillCmd.AddCommand(
		newSkillAddCmd(),
		newSkillRemoveCmd(),
		newSkillUpdateCmd(),
		newSkillListCmd(),
	)
	return skillCmd
}

func newSkillAddCmd() *cobra.Command {
	var global bool
	var agents []string
	var all bool

	cmd := &cobra.Command{
		Use:   "add <owner/skill>",
		Short: "Install a skill from skills.sh",
		Example: `  usher skill add supabase/agent-skills
  usher skill add supabase/agent-skills --global
  usher skill add supabase/agent-skills --agent claude --agent cursor`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := actions.NewSkillActions(d.cfgWriter, makeSyncFn(), d.out, d.prompt)
			return a.Add(args[0], global, cmd.Flags().Changed("global"), agents, all)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Install globally (~/.config/...)")
	cmd.Flags().StringArrayVarP(&agents, "agent", "a", nil, "Target specific agents (e.g. claude, cursor)")
	cmd.Flags().BoolVar(&all, "all", false, "Install all skills without prompts")
	return cmd
}

func newSkillRemoveCmd() *cobra.Command {
	var global bool
	var agents []string

	cmd := &cobra.Command{
		Use:   "remove [skill-name]",
		Short: "Remove an installed skill",
		Example: `  usher skill remove           # interactive multi-select
  usher skill remove supabase  # remove by name`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			a := actions.NewSkillActions(d.cfgWriter, makeSyncFn(), d.out, d.prompt)
			return a.Remove(name, global, cmd.Flags().Changed("global"), agents)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Remove from global scope")
	cmd.Flags().StringArrayVarP(&agents, "agent", "a", nil, "Target specific agents")
	return cmd
}

func newSkillUpdateCmd() *cobra.Command {
	var global bool
	var agents []string
	var all bool

	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update installed skills to latest versions",
		Example: `  usher skill update
  usher skill update --global`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			a := actions.NewSkillActions(d.cfgWriter, makeSyncFn(), d.out, d.prompt)
			return a.Update(global, cmd.Flags().Changed("global"), agents, all)
		},
	}
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Update global skills")
	cmd.Flags().StringArrayVarP(&agents, "agent", "a", nil, "Target specific agents")
	cmd.Flags().BoolVar(&all, "all", false, "Update all without prompts")
	return cmd
}

func newSkillListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			a := actions.NewSkillActions(d.cfgWriter, makeSyncFn(), d.out, d.prompt)
			return a.List()
		},
	}
}
