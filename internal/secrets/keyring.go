package secrets

import (
	"errors"
	"fmt"

	"github.com/99designs/keyring"
)

const (
	serviceName = "igpostercli"
)

var errTokenNotFound = errors.New("access token not found")

func OpenKeyring() (keyring.Keyring, error) {
	kr, err := keyring.Open(keyring.Config{
		ServiceName: serviceName,
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
			keyring.KWalletBackend,
			keyring.PassBackend,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}

	return kr, nil
}

func tokenKey(profile string) string {
	return fmt.Sprintf("profile:%s", profile)
}

func SetAccessToken(profile string, token string) error {
	kr, err := OpenKeyring()
	if err != nil {
		return err
	}

	item := keyring.Item{
		Key:   tokenKey(profile),
		Data:  []byte(token),
		Label: "igpostercli access token",
	}

	if err := kr.Set(item); err != nil {
		return fmt.Errorf("store access token: %w", err)
	}

	return nil
}

func GetAccessToken(profile string) (string, bool, error) {
	kr, err := OpenKeyring()
	if err != nil {
		return "", false, err
	}

	item, err := kr.Get(tokenKey(profile))
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", false, nil
		}

		return "", false, fmt.Errorf("get access token: %w", err)
	}

	return string(item.Data), true, nil
}

func DeleteAccessToken(profile string) (bool, error) {
	kr, err := OpenKeyring()
	if err != nil {
		return false, err
	}

	if err := kr.Remove(tokenKey(profile)); err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return false, errTokenNotFound
		}

		return false, fmt.Errorf("delete access token: %w", err)
	}

	return true, nil
}

func ErrTokenNotFound() error {
	return errTokenNotFound
}
