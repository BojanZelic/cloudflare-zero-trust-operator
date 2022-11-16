package cftypes

import "github.com/cloudflare/cloudflare-go"

type ExtendedServiceToken struct {
	cloudflare.AccessServiceToken
	ClientSecret string
}
