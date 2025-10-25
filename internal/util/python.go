package util

import "fmt"

// findPythonExecutable finds a suitable Python executable
func FindPythonExecutable() (string, error) {
	// Try python3 first, then python
	candidates := []string{"python3", "python"}

	for _, candidate := range candidates {
		if err := CheckCommandExists(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no Python executable found (tried: %v)", candidates)
}
