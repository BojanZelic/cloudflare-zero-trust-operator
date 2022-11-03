package config

import (
	"errors"
	"os"
)

var CLOUDFLARE_API_EMAIL string
var CLOUDFLARE_API_KEY string
var CLOUDFLARE_API_TOKEN string
var CLOUDFLARE_ACCOUNT_ID string

func ReadFromEnv() {
	CLOUDFLARE_API_EMAIL = os.Getenv("CLOUDFLARE_API_EMAIL")
	CLOUDFLARE_API_KEY = os.Getenv("CLOUDFLARE_API_KEY")
	CLOUDFLARE_API_TOKEN = os.Getenv("CLOUDFLARE_API_TOKEN")
	CLOUDFLARE_ACCOUNT_ID = os.Getenv("CLOUDFLARE_ACCOUNT_ID")
}

func IsValid() (bool, error) {
	if CLOUDFLARE_ACCOUNT_ID == "" {
		return false, errors.New("CLOUDFLARE_ACCOUNT_ID needs to be set")
	}

	if CLOUDFLARE_API_TOKEN == "" && (CLOUDFLARE_API_EMAIL == "" && CLOUDFLARE_API_KEY == "") {
		return false, errors.New("One of CLOUDFLARE_API_TOKEN or (CLOUDFLARE_API_EMAIL and CLOUDFLARE_API_KEY) needs to be set")
	}

	return true, nil
}
