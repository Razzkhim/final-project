package config

import "os"

var (
	JWTKey = []byte(getJWTKey())
)

func getJWTKey() string {
	key := os.Getenv("JWT_KEY")
	if key == "" {
		key = "super_duper_secret_key"
	}

	return key
}
