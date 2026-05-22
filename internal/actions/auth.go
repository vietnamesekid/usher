package actions

import (
	"fmt"
	"time"

	"github.com/vietnamesekid/usher/internal/config"
	"github.com/vietnamesekid/usher/internal/keychain"
	"github.com/vietnamesekid/usher/internal/ui"
)

type provider struct {
	name    string
	keyType string
	label   string
	url     string
}

var providers = []provider{
	{
		name:    "anthropic",
		keyType: "api_key",
		label:   "Anthropic (Claude Code)",
		url:     "https://console.anthropic.com/settings/keys",
	},
	{
		name:    "google",
		keyType: "api_key",
		label:   "Google (Gemini CLI)",
		url:     "https://aistudio.google.com/app/apikey",
	},
	{
		name:    "openai",
		keyType: "api_key",
		label:   "OpenAI (Codex CLI)",
		url:     "https://platform.openai.com/api-keys",
	},
}

type AuthActions struct {
	cfgLoader *config.Loader
	cfgWriter *config.Writer
	kc        keychain.Keychain
	out       *ui.Output
	prompt    *ui.Prompt
}

func NewAuthActions(
	cfgLoader *config.Loader,
	cfgWriter *config.Writer,
	kc keychain.Keychain,
	out *ui.Output,
	prompt *ui.Prompt,
) *AuthActions {
	return &AuthActions{
		cfgLoader: cfgLoader,
		cfgWriter: cfgWriter,
		kc:        kc,
		out:       out,
		prompt:    prompt,
	}
}

func (a *AuthActions) Setup() error {
	for _, p := range providers {
		if !a.prompt.AskConfirm(fmt.Sprintf("Set up %s?", p.label)) {
			continue
		}
		if p.url != "" {
			a.out.Info(fmt.Sprintf("Get your API key at: %s", p.url))
		}
		token := a.prompt.AskSecret(fmt.Sprintf("Enter API key for %s:", p.label))
		if token == "" {
			a.out.Warning(fmt.Sprintf("Skipping %s (no key entered)", p.label))
			continue
		}
		kcKey := keychain.AuthKey(p.name, p.keyType)
		if err := a.kc.Set(kcKey, token); err != nil {
			return fmt.Errorf("storing key for %s: %w", p.name, err)
		}
		entry := config.AuthEntry{
			Provider: p.name,
			KeyRef:   kcKey,
			AddedAt:  time.Now(),
		}
		if err := a.cfgWriter.AddAuthEntry(entry); err != nil {
			return fmt.Errorf("saving auth entry: %w", err)
		}
		a.out.Success(fmt.Sprintf("Saved API key for %s", p.label))
	}
	return nil
}

func (a *AuthActions) Revoke(providerName string) error {
	global, _, err := a.cfgLoader.Load()
	if err != nil {
		return err
	}

	var found config.AuthEntry
	for _, e := range global.Auth {
		if e.Provider == providerName {
			found = e
			break
		}
	}
	if found.KeyRef == "" {
		return fmt.Errorf("provider %q is not configured", providerName)
	}

	if err := a.kc.Delete(found.KeyRef); err != nil {
		a.out.Warning(fmt.Sprintf("could not delete keychain entry: %v", err))
	}

	a.out.Success(fmt.Sprintf("Revoked credentials for %q", providerName))
	return nil
}
