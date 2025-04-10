package errorz

import "fmt"

func WrapError(moduleName, action string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: failed to %s: %w", moduleName, action, err)
}
