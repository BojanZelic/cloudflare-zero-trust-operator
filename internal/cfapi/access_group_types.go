package cfapi

import (
	cloudflare "github.com/cloudflare/cloudflare-go"
)

func NewAccessGroupEmail(email string) cloudflare.AccessGroupEmail {
	return cloudflare.AccessGroupEmail{Email: struct {
		Email string "json:\"email\""
	}{
		Email: email,
	}}
}

func NewAccessGroupEmailDomains(domain string) cloudflare.AccessGroupEmailDomain {
	return cloudflare.AccessGroupEmailDomain{EmailDomain: struct {
		Domain string "json:\"domain\""
	}{
		Domain: domain,
	}}
}

func NewAccessGroupIP(ip string) cloudflare.AccessGroupIP {
	return cloudflare.AccessGroupIP{
		IP: struct {
			IP string "json:\"ip\""
		}{
			IP: ip,
		},
	}
}

func NewAccessGroupServiceToken(token string) cloudflare.AccessGroupServiceToken {
	return cloudflare.AccessGroupServiceToken{
		ServiceToken: struct {
			ID string "json:\"token_id\""
		}{
			ID: token,
		},
	}
}

func NewAccessGroupAccessGroup(id string) cloudflare.AccessGroupAccessGroup {
	return cloudflare.AccessGroupAccessGroup{
		Group: struct {
			ID string "json:\"id\""
		}{
			ID: id,
		},
	}
}

func NewAccessGroupAnyValidServiceToken() cloudflare.AccessGroupAnyValidServiceToken {
	return cloudflare.AccessGroupAnyValidServiceToken{
		AnyValidServiceToken: struct{}{},
	}
}
