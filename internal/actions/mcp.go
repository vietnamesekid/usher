package actions

import (
	"fmt"

	"github.com/vietnamesekid/usher/internal/config"
	"github.com/vietnamesekid/usher/internal/keychain"
	"github.com/vietnamesekid/usher/internal/registry"
	"github.com/vietnamesekid/usher/internal/ui"
)

type MCPActions struct {
	cfgLoader *config.Loader
	cfgWriter *config.Writer
	reg       registry.Registry
	kc        keychain.Keychain
	syncFn    func(global config.Config, project config.ProjectConfig) error
	out       *ui.Output
	prompt    *ui.Prompt
}

func NewMCPActions(
	cfgLoader *config.Loader,
	cfgWriter *config.Writer,
	reg registry.Registry,
	kc keychain.Keychain,
	syncFn func(config.Config, config.ProjectConfig) error,
	out *ui.Output,
	prompt *ui.Prompt,
) *MCPActions {
	return &MCPActions{
		cfgLoader: cfgLoader,
		cfgWriter: cfgWriter,
		reg:       reg,
		kc:        kc,
		syncFn:    syncFn,
		out:       out,
		prompt:    prompt,
	}
}

func (a *MCPActions) Add(name string) error {
	regEntry, err := a.reg.GetMCP(name)
	if err != nil {
		return err
	}

	global, _, err := a.cfgLoader.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	instanceKey := name
	existing, alreadyAdded := global.MCPServers[name]

	if alreadyAdded && len(existing.Instances) > 0 {
		choice := a.prompt.AskSelect(
			fmt.Sprintf("%q is already configured. What would you like to do?", name),
			[]string{"Update token", "Add new instance"},
		)
		if choice == "Add new instance" {
			instName := a.prompt.AskString("Instance name (e.g. prod, staging):", "")
			if instName == "" {
				return fmt.Errorf("instance name is required")
			}
			instanceKey = name + "-" + instName
		}
	}

	// MCPs with no auth (e.g. filesystem) skip the token prompt entirely.
	noAuth := regEntry.Auth.EnvVar == ""

	var newInstance config.MCPInstance
	if noAuth {
		newInstance = config.MCPInstance{
			Name:    instanceKey,
			Auth:    config.AuthRef{Type: "none", Key: ""},
			Enabled: true,
		}
	} else {
		if regEntry.Auth.URL != "" {
			a.out.Info(fmt.Sprintf("Get your token at: %s", regEntry.Auth.URL))
		}
		token := a.prompt.AskSecret(fmt.Sprintf("Enter %s (%s):", regEntry.Auth.EnvVar, name))
		if token == "" {
			return fmt.Errorf("token is required")
		}
		kcKey := keychain.MCPTokenKey(instanceKey)
		if err := a.kc.Set(kcKey, token); err != nil {
			return fmt.Errorf("storing token in keychain: %w", err)
		}
		newInstance = config.MCPInstance{
			Name:    instanceKey,
			Auth:    config.AuthRef{Type: "keychain", Key: kcKey},
			Enabled: true,
		}
	}

	var instances []config.MCPInstance
	if alreadyAdded {
		// Replace instance with same key, or append.
		replaced := false
		for _, inst := range existing.Instances {
			if inst.Name == instanceKey {
				instances = append(instances, newInstance)
				replaced = true
			} else {
				instances = append(instances, inst)
			}
		}
		if !replaced {
			instances = append(instances, newInstance)
		}
	} else {
		instances = []config.MCPInstance{newInstance}
	}

	entry := config.MCPEntry{Instances: instances}
	if err := a.cfgWriter.AddMCPEntry(name, entry); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	global, project, err := a.cfgLoader.LoadBoth()
	if err != nil {
		return err
	}
	if err := a.syncFn(global, project); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	a.out.Success(fmt.Sprintf("Added MCP %q (instance: %s)", name, instanceKey))
	return nil
}

func (a *MCPActions) Remove(name string) error {
	global, _, err := a.cfgLoader.Load()
	if err != nil {
		return err
	}

	entry, ok := global.MCPServers[name]
	if !ok {
		return fmt.Errorf("MCP %q is not configured", name)
	}

	for _, inst := range entry.Instances {
		if err := a.kc.Delete(inst.Auth.Key); err != nil {
			a.out.Warning(fmt.Sprintf("could not delete keychain key %q: %v", inst.Auth.Key, err))
		}
	}

	if err := a.cfgWriter.RemoveMCPEntry(name); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	global, project, err := a.cfgLoader.LoadBoth()
	if err != nil {
		return err
	}
	if err := a.syncFn(global, project); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	a.out.Success(fmt.Sprintf("Removed MCP %q", name))
	return nil
}

func (a *MCPActions) List() error {
	global, _, err := a.cfgLoader.Load()
	if err != nil {
		return err
	}

	if len(global.MCPServers) == 0 {
		a.out.Info("No MCP servers configured. Run: usher mcp add <name>")
		return nil
	}

	headers := []string{"NAME", "INSTANCE", "TOKEN"}
	var rows [][]string
	for serverName, entry := range global.MCPServers {
		for _, inst := range entry.Instances {
			tokenStatus := "missing"
			if ok, _ := a.kc.Exists(inst.Auth.Key); ok {
				tokenStatus = "ok"
			}
			rows = append(rows, []string{serverName, inst.Name, tokenStatus})
		}
	}
	a.out.Table(headers, rows)
	return nil
}

func (a *MCPActions) ListAvailable() error {
	entries := a.reg.ListMCPs()
	headers := []string{"NAME", "DESCRIPTION", "TAGS"}
	var rows [][]string
	for _, e := range entries {
		tags := ""
		if len(e.Tags) > 0 {
			tags = fmt.Sprintf("%v", e.Tags)
		}
		rows = append(rows, []string{e.Name, e.Description, tags})
	}
	a.out.Table(headers, rows)
	return nil
}
