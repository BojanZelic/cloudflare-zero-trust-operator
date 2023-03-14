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

func NewAccessGroupGSuite(email string, identityProviderID string) cloudflare.AccessGroupGSuite {
	return cloudflare.AccessGroupGSuite{
		Gsuite: struct {
			Email              string "json:\"email\""
			IdentityProviderID string "json:\"identity_provider_id\""
		}{
			Email:              email,
			IdentityProviderID: identityProviderID,
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

func NewAccessGroupGeo(country string) cloudflare.AccessGroupGeo {
	return cloudflare.AccessGroupGeo{
		Geo: struct {
			CountryCode string "json:\"country_code\""
		}{
			CountryCode: country,
		},
	}
}

func NewAccessGroupEveryone() cloudflare.AccessGroupEveryone {
	return cloudflare.AccessGroupEveryone{
		Everyone: struct{}{},
	}
}

func NewAccessGroupCertificate() cloudflare.AccessGroupCertificate {
	return cloudflare.AccessGroupCertificate{
		Certificate: struct{}{},
	}
}

func NewAccessGroupLoginMethod(id string) cloudflare.AccessGroupLoginMethod {
	return cloudflare.AccessGroupLoginMethod{
		LoginMethod: struct {
			ID string "json:\"id\""
		}{
			ID: id,
		},
	}
}

func NewAccessGroupOktaGroup(name string, identityProviderID string) cloudflare.AccessGroupOkta {
	return cloudflare.AccessGroupOkta{
		Okta: struct {
			Name               string "json:\"name\""
			IdentityProviderID string "json:\"identity_provider_id\""
		}{
			Name:               name,
			IdentityProviderID: identityProviderID,
		},
	}
}
