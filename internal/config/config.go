package config

import (
	"errors"
	"os"
)

var CLOUDFLARE_API_EMAIL string
var CLOUDFLARE_API_KEY string
var CLOUDFLARE_API_TOKEN string

func ReadFromEnv() {
	CLOUDFLARE_API_EMAIL = os.Getenv("CLOUDFLARE_API_EMAIL")
	CLOUDFLARE_API_KEY = os.Getenv("CLOUDFLARE_API_KEY")
	CLOUDFLARE_API_TOKEN = os.Getenv("CLOUDFLARE_API_TOKEN")
}

func IsValid() (bool, error) {
	if CLOUDFLARE_API_TOKEN != "" {
		return true, nil
	}

	if CLOUDFLARE_API_EMAIL != "" && CLOUDFLARE_API_KEY != "" {
		return true, nil
	}

	return false, errors.New("One of CLOUDFLARE_API_TOKEN or (CLOUDFLARE_API_EMAIL and CLOUDFLARE_API_KEY) needs to be set")
}
