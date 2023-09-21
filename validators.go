package uinput

import (
	"errors"
	"fmt"
	"os"
)

func ValidateDevicePath(path string) error {
	if path == "" {
		return errors.New("device path must not be empty")
	}
	_, err := os.Stat(path)
	return err
}

func ValidateUinputName(name string) error {
	if len(name) == 0 {
		return errors.New("device name may not be empty")
	}
	if len(name) > UinputMaxNameSize {
		return fmt.Errorf("device name %s is too long (maximum of %d characters allowed)", name, UinputMaxNameSize)
	}
	return nil
}
