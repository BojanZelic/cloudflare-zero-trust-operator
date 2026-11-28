package cfapi

import (
	cloudflare "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/zero_trust"
)

func NewAccessGroupEmail(email string) zero_trust.EmailRuleParam {
	return zero_trust.EmailRuleParam{
		Email: cloudflare.F(zero_trust.EmailRuleEmailParam{
			Email: cloudflare.F(email),
		}),
	}
}

func NewAccessGroupEmailDomains(domain string) zero_trust.DomainRuleParam {
	return zero_trust.DomainRuleParam{
		EmailDomain: cloudflare.F(zero_trust.DomainRuleEmailDomainParam{
			Domain: cloudflare.F(domain),
		}),
	}
}

func NewAccessGroupIP(ip string) zero_trust.IPRuleParam {
	return zero_trust.IPRuleParam{
		IP: cloudflare.F(zero_trust.IPRuleIPParam{
			IP: cloudflare.F(ip),
		}),
	}
}

func NewAccessGroupGSuite(email string, identityProviderID string) zero_trust.GSuiteGroupRuleParam {
	return zero_trust.GSuiteGroupRuleParam{
		GSuite: cloudflare.F(zero_trust.GSuiteGroupRuleGSuiteParam{
			Email:              cloudflare.F(email),
			IdentityProviderID: cloudflare.F(identityProviderID),
		}),
	}
}

func NewAccessGroupServiceToken(token string) zero_trust.ServiceTokenRuleParam {
	return zero_trust.ServiceTokenRuleParam{
		ServiceToken: cloudflare.F(zero_trust.ServiceTokenRuleServiceTokenParam{
			TokenID: cloudflare.F(token),
		}),
	}
}

func NewAccessGroupAccessGroup(id string) zero_trust.GroupRuleParam {
	return zero_trust.GroupRuleParam{
		Group: cloudflare.F(zero_trust.GroupRuleGroupParam{
			ID: cloudflare.F(id),
		}),
	}
}

func NewAccessGroupAnyValidServiceToken() zero_trust.AnyValidServiceTokenRuleParam {
	return zero_trust.AnyValidServiceTokenRuleParam{
		AnyValidServiceToken: cloudflare.F(zero_trust.AnyValidServiceTokenRuleAnyValidServiceTokenParam{}),
	}
}

func NewAccessGroupGeo(country string) zero_trust.CountryRuleParam {
	return zero_trust.CountryRuleParam{
		Geo: cloudflare.F(zero_trust.CountryRuleGeoParam{
			CountryCode: cloudflare.F(country),
		}),
	}
}

func NewAccessGroupEveryone() zero_trust.EveryoneRuleParam {
	return zero_trust.EveryoneRuleParam{
		Everyone: cloudflare.F(zero_trust.EveryoneRuleEveryoneParam{}),
	}
}

func NewAccessGroupCertificate() zero_trust.CertificateRuleParam {
	return zero_trust.CertificateRuleParam{
		Certificate: cloudflare.F(zero_trust.CertificateRuleCertificateParam{}),
	}
}

func NewAccessGroupCommonName(commonName string) zero_trust.AccessRuleAccessCommonNameRuleParam {
	return zero_trust.AccessRuleAccessCommonNameRuleParam{
		CommonName: cloudflare.F(zero_trust.AccessRuleAccessCommonNameRuleCommonNameParam{
			CommonName: cloudflare.F(commonName),
		}),
	}
}

func NewAccessGroupLoginMethod(id string) zero_trust.AccessRuleAccessLoginMethodRuleParam {
	return zero_trust.AccessRuleAccessLoginMethodRuleParam{
		LoginMethod: cloudflare.F(zero_trust.AccessRuleAccessLoginMethodRuleLoginMethodParam{
			ID: cloudflare.F(id),
		}),
	}
}

func NewAccessGroupOktaGroup(name string, identityProviderID string) zero_trust.OktaGroupRuleParam {
	return zero_trust.OktaGroupRuleParam{
		Okta: cloudflare.F(zero_trust.OktaGroupRuleOktaParam{
			Name:               cloudflare.F(name),
			IdentityProviderID: cloudflare.F(identityProviderID),
		}),
	}
}

func NewAccessGroupOIDCClaim(name string, value string, identityProviderID string) zero_trust.AccessRuleAccessOIDCClaimRuleParam {
	return zero_trust.AccessRuleAccessOIDCClaimRuleParam{
		OIDC: cloudflare.F(zero_trust.AccessRuleAccessOIDCClaimRuleOIDCParam{
			ClaimName:          cloudflare.F(name),
			ClaimValue:         cloudflare.F(value),
			IdentityProviderID: cloudflare.F(identityProviderID),
		}),
	}
}
