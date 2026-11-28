package cftypes

import (
	"time"

	"github.com/cloudflare/cloudflare-go/v6/zero_trust"
)

type AccessApplicationType = zero_trust.ApplicationType

const (
	AccessApplicationTypeSelfHosted AccessApplicationType = zero_trust.ApplicationTypeSelfHosted
)

type AccessApplication struct {
	ID                      string
	Name                    string
	Domain                  string
	Type                    AccessApplicationType
	AppLauncherVisible      *bool
	AllowedIdps             []string
	AutoRedirectToIdentity  *bool
	SessionDuration         string
	EnableBindingCookie     *bool
	HttpOnlyCookieAttribute *bool
	LogoURL                 string
	CreatedAt               *time.Time
	UpdatedAt               *time.Time
	PolicyIDs               []string
}

type AccessGroup struct {
	ID        string
	Name      string
	Include   []interface{}
	Exclude   []interface{}
	Require   []interface{}
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type AccessPolicy struct {
	ID                           string
	Name                         string
	Precedence                   int
	Decision                     string
	Include                      []interface{}
	Exclude                      []interface{}
	Require                      []interface{}
	Reusable                     bool
	AppCount                     int64
	IsolationRequired            *bool
	SessionDuration              *string
	PurposeJustificationRequired *bool
	PurposeJustificationPrompt   *string
	ApprovalRequired             *bool
	CreatedAt                    *time.Time
	UpdatedAt                    *time.Time
}

type AccessServiceToken struct {
	ID        string
	Name      string
	ClientID  string
	Duration  string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	ExpiresAt *time.Time
}
