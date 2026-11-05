package skills

import (
	"errors"
	"fmt"
)

// Limits from the Agent Skills format.
const (
	maxNameLength        = 64
	maxDescriptionLength = 1024
)

// validateName enforces the Agent Skills naming rule, which also keeps the
// name safe to use as a folder name.
func validateName(name string) error {
	if name == "" {
		return errors.New("frontmatter is missing the required name")
	}
	if len(name) > maxNameLength {
		return fmt.Errorf("name %q is longer than %d characters", name, maxNameLength)
	}
	for _, r := range name {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return fmt.Errorf("name %q may only contain lowercase letters, digits and hyphens", name)
		}
	}
	return nil
}

// validateDescription requires a description the model can match tasks on.
func validateDescription(description string) error {
	if description == "" {
		return errors.New("frontmatter is missing the required description")
	}
	if len(description) > maxDescriptionLength {
		return fmt.Errorf("description is longer than %d characters", maxDescriptionLength)
	}
	return nil
}
