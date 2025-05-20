package cfapi

import (
	"context"

	gin "github.com/onsi/ginkgo/v2"
)

//
//
//

// Helps to trace created CloudFlare ressources while a controllers run. Its main purpose is to allow a strict cleanup of created ressources.
type InsertedCFRessourcesTracer struct {
	accessGroupCF_UUIDs          []string
	accessServiceTokenCF_UUIDs   []string
	accessReusablePolicyCF_UUIDs []string
	accessApplicationCF_UUIDs    []string
}

func (a *InsertedCFRessourcesTracer) ResetCFUUIDs() {
	a.accessGroupCF_UUIDs = []string{}
	a.accessServiceTokenCF_UUIDs = []string{}
	a.accessReusablePolicyCF_UUIDs = []string{}
	a.accessApplicationCF_UUIDs = []string{}
}

// Will remove all tracked ressources from CloudFlare
func (a *InsertedCFRessourcesTracer) UninstallFromCF(api *API) {
	ctx := context.Background()

	//
	// @dev cleanup order is important, since some depends on others !
	//

	if len(a.accessApplicationCF_UUIDs) > 0 {
		gin.By("Removing all created access apps in this test set")
		for _, appID := range a.accessApplicationCF_UUIDs {
			_ = api.DeleteGenericAccessApplication(ctx, appID)
		}
	}

	if len(a.accessReusablePolicyCF_UUIDs) > 0 {
		gin.By("Removing all created reusable policies in this test set")
		for _, rpID := range a.accessReusablePolicyCF_UUIDs {
			_ = api.DeleteAccessReusablePolicy(ctx, rpID)
		}
	}

	if len(a.accessGroupCF_UUIDs) > 0 {
		gin.By("Removing all created access groups in this test set")
		for _, groupID := range a.accessGroupCF_UUIDs {
			_ = api.DeleteAccessGroup(ctx, groupID)
		}
	}

	if len(a.accessServiceTokenCF_UUIDs) > 0 {
		gin.By("Removing all created service tokens in this test set")
		for _, tokenID := range a.accessServiceTokenCF_UUIDs {
			_ = api.DeleteAccessServiceToken(ctx, tokenID)
		}
	}
}

//
//
//

func (a *InsertedCFRessourcesTracer) GroupInserted(id string) {
	a.traceInsert(&a.accessGroupCF_UUIDs, id)
}
func (a *InsertedCFRessourcesTracer) ServiceTokenInserted(id string) {
	a.traceInsert(&a.accessServiceTokenCF_UUIDs, id)
}
func (a *InsertedCFRessourcesTracer) ReusablePolicyInserted(id string) {
	a.traceInsert(&a.accessReusablePolicyCF_UUIDs, id)
}
func (a *InsertedCFRessourcesTracer) ApplicationInserted(id string) {
	a.traceInsert(&a.accessApplicationCF_UUIDs, id)
}

//
//

func (a *InsertedCFRessourcesTracer) GroupDeleted(id string) {
	a.traceDelete(&a.accessGroupCF_UUIDs, id)
}
func (a *InsertedCFRessourcesTracer) ServiceTokenDeleted(id string) {
	a.traceDelete(&a.accessServiceTokenCF_UUIDs, id)
}
func (a *InsertedCFRessourcesTracer) ReusablePolicyDeleted(id string) {
	a.traceDelete(&a.accessReusablePolicyCF_UUIDs, id)
}
func (a *InsertedCFRessourcesTracer) ApplicationDeleted(id string) {
	a.traceDelete(&a.accessApplicationCF_UUIDs, id)
}

//
//
//

func (a *InsertedCFRessourcesTracer) traceInsert(store *[]string, idInserted string) {
	*store = append(*store, idInserted)
}

func (a *InsertedCFRessourcesTracer) traceDelete(store *[]string, idRemoved string) {
	s := *store //nolint:varnamelen
	n := 0      //nolint:varnamelen
	for _, v := range s {
		if v != idRemoved {
			s[n] = v
			n++
		}
	}
	*store = s[:n]
}
