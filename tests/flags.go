package tests

import "flag"

//
// Tests configuration
//

type TestFlagsStruct struct {
	// Would skip any tests regarding unconfigured `one-app` applications types on the targeted CloudFlare Account (`warp`, `app_launcher`...)
	SkipUnconfiguredOneApp *bool
}

var TestFlags TestFlagsStruct = TestFlagsStruct{
	SkipUnconfiguredOneApp: flag.Bool(
		"skipUnconfiguredOneApp", false,
		"Would skip any tests regarding unconfigured `one-app` applications types on the targeted CloudFlare Account (`warp`, `app_launcher`...)",
	),
}
