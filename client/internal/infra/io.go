package infra

import (
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
)

func StoreToken(token uuid.UUID) error {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to retrieve config root on client machine: %w", err)
	}

	appConfigDir := fmt.Sprintf("%s/%s", configRoot, "chat-client")
	err = os.MkdirAll(appConfigDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create app config dir: %w", err)
	}

	tokenFilePath := fmt.Sprintf("%s/%s", appConfigDir, "token")
	err = os.WriteFile(tokenFilePath, []byte(token.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write token to file: %w", err)
	}

	return nil
}

func Token() (string, error) {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve config root on client machine: %w", err)
	}

	tokenFilePath := fmt.Sprintf("%s/%s/%s", configRoot, "chat-client", "token")
	data, err := os.ReadFile(tokenFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}

		return "", fmt.Errorf("failed to read token file: %w", err)
	}

	return string(data), nil
}
