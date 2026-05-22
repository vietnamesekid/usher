package config

import "fmt"

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func Validate(cfg Config) []ValidationError {
	var errs []ValidationError

	if cfg.Version == "" {
		errs = append(errs, ValidationError{"version", "is required"})
	}

	for name, entry := range cfg.MCPServers {
		if len(entry.Instances) == 0 {
			errs = append(errs, ValidationError{
				fmt.Sprintf("mcpServers.%s", name),
				"must have at least one instance",
			})
			continue
		}
		for i, inst := range entry.Instances {
			field := fmt.Sprintf("mcpServers.%s.instances[%d]", name, i)
			if inst.Name == "" {
				errs = append(errs, ValidationError{field + ".name", "is required"})
			}
			if inst.Auth.Type != "none" && inst.Auth.Key == "" {
				errs = append(errs, ValidationError{field + ".auth.key", "is required"})
			}
			if inst.Auth.Type != "keychain" && inst.Auth.Type != "env" && inst.Auth.Type != "none" {
				errs = append(errs, ValidationError{field + ".auth.type", `must be "keychain", "env", or "none"`})
			}
		}
	}

	for name, entry := range cfg.Skills {
		if entry.Version == "" {
			errs = append(errs, ValidationError{
				fmt.Sprintf("skills.%s.version", name),
				"is required",
			})
		}
	}

	return errs
}
