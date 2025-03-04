package util

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

func MakeAdminKey(jwtSecret string) (string, error) {
	adminKey, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "service_role",
	}).SignedString([]byte(jwtSecret))
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate admin key")
	}

	return adminKey, nil
}

func MakeAnonKey(jwtSecret string) (string, error) {
	anonKey, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "anon",
	}).SignedString([]byte(jwtSecret))
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate anon key")
	}

	return anonKey, nil
}
