package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/vietnamesekid/usher/internal/config"
	"github.com/vietnamesekid/usher/internal/keychain"
	"github.com/vietnamesekid/usher/internal/registry"
	synctypes "github.com/vietnamesekid/usher/internal/sync"
	"github.com/vietnamesekid/usher/internal/ui"
)

// deps holds all initialized dependencies, shared across subcommands.
type deps struct {
	cfgLoader *config.Loader
	cfgWriter *config.Writer
	reg       registry.Registry
	kc        keychain.Keychain
	out       *ui.Output
	prompt    *ui.Prompt
}

var (
	d              deps
	configOverride string
)

func NewRootCmd(version, commit, date string) *cobra.Command {
	root := &cobra.Command{
		Use:     "usher",
		Short:   "Aggregator and installer for AI coding tools",
		Version: fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
		Long: `Usher manages MCP servers and skills across AI coding tools
(Claude Code, Gemini CLI, Codex, Cursor) from a single config.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// setup initializes deps itself — skip PersistentPreRunE.
			if cmd.Name() == "setup" {
				return initDepsMinimal()
			}
			return initDeps(configOverride)
		},
	}

	root.PersistentFlags().StringVar(&configOverride, "config", "", "path to config file (default: ~/.usher/config.json)")

	root.AddCommand(
		newSetupCmd(),
		newSyncCmd(),
		newMCPCmd(),
		newSkillCmd(),
		newAuthCmd(),
		newDoctorCmd(),
	)

	return root
}

// initDepsMinimal initializes only UI deps (used by setup before config exists).
func initDepsMinimal() error {
	d.out = ui.NewOutput()
	d.prompt = ui.NewPrompt()
	return nil
}

func initDeps(configPath string) error {
	d.out = ui.NewOutput()
	d.prompt = ui.NewPrompt()
	d.cfgLoader = config.NewLoader(configPath)

	globalPath := d.cfgLoader.GlobalPath()
	d.cfgWriter = config.NewWriter(globalPath)

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("finding home dir: %w", err)
	}

	cacheDir := filepath.Join(home, ".usher", "registry")
	cache := registry.NewCache(cacheDir, 24*time.Hour)
	d.reg = registry.NewWithCache(cache)

	usherDir := filepath.Join(home, ".usher")
	d.kc = keychain.New(usherDir)

	return nil
}

// makeSyncFn returns a sync function suitable for injection into actions.
func makeSyncFn() func(global config.Config, project config.ProjectConfig) error {
	return func(global config.Config, project config.ProjectConfig) error {
		syncer := synctypes.New(
			global,
			project,
			d.reg,
			d.kc,
			d.prompt.AskSecret,
			d.out,
		)
		return syncer.Run()
	}
}
