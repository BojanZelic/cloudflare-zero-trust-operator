package cfapi

import (
	"encoding/json"
	"reflect"

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

func AcessGroupEmailEqual(first cloudflare.AccessGroup, second cloudflare.AccessGroup) bool {
	v1, _ := json.Marshal(first.Include)
	v2, _ := json.Marshal(second.Include)
	if !reflect.DeepEqual(v1, v2) {
		return false
	}
	v1, _ = json.Marshal(first.Exclude)
	v2, _ = json.Marshal(second.Exclude)
	if !reflect.DeepEqual(v1, v2) {
		return false
	}
	v1, _ = json.Marshal(first.Require)
	v2, _ = json.Marshal(second.Require)
	if !reflect.DeepEqual(v1, v2) {
		return false
	}

	return true
}

func AccessPoliciesEqual(first []cloudflare.AccessPolicy, second []cloudflare.AccessPolicy) bool {

	if len(first) != len(second) {
		return false
	}

	for i := range first {
		if first[i].Name != second[i].Name {
			return false
		}
		if first[i].Precedence != second[i].Precedence {
			return false
		}
		v1, _ := json.Marshal(first[i].Include)
		v2, _ := json.Marshal(second[i].Include)
		if !reflect.DeepEqual(v1, v2) {
			return false
		}
		v1, _ = json.Marshal(first[i].Exclude)
		v2, _ = json.Marshal(second[i].Exclude)
		if !reflect.DeepEqual(v1, v2) {
			return false
		}
		v1, _ = json.Marshal(first[i].Require)
		v2, _ = json.Marshal(second[i].Require)
		if !reflect.DeepEqual(v1, v2) {
			return false
		}
	}

	return true
}
