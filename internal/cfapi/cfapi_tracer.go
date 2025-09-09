package cfapi

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	gin "github.com/onsi/ginkgo/v2"
)

//
//
//

// key: UUID, value: description helper
type CfUUIDStore map[string]string

func (store *CfUUIDStore) TraceInsert(cfUUID string, resourceDescription string) {
	(*store)[cfUUID] = resourceDescription
}

func (store *CfUUIDStore) TraceDelete(cfUUID string) {
	delete(*store, cfUUID)
}

// Helps to trace created CloudFlare ressources while a controllers run.
// Its main purpose is to allow a strict cleanup of created ressources.
type CloudflareResourceCreationTracer struct {
	AccessGroups           CfUUIDStore
	AccessServiceTokens    CfUUIDStore
	AccessReusablePolicies CfUUIDStore
	AccessApplications     CfUUIDStore
}

func (tracer *CloudflareResourceCreationTracer) ResetStores() {
	tracer.AccessGroups = CfUUIDStore{}
	tracer.AccessServiceTokens = CfUUIDStore{}
	tracer.AccessReusablePolicies = CfUUIDStore{}
	tracer.AccessApplications = CfUUIDStore{}
}

// Will attempt to remove all stored ressources from CloudFlare
//
// nolint:cyclop
func (tracer *CloudflareResourceCreationTracer) UninstallFromCF(api *API) []error {
	errs := []error{}

	//
	// @dev cleanup order is important, since some depends on others !
	//

	//
	tryDeleteAllResourcesFromStore(
		"Removing all created access applications in this test set",
		&tracer.AccessApplications,
		&errs,
		func(ctx context.Context, UUID string) error {
			return api.DeleteGenericAccessApplication(ctx, UUID)
		},
	)

	//
	tryDeleteAllResourcesFromStore(
		"Removing all created access reusable policies in this test set",
		&tracer.AccessReusablePolicies,
		&errs,
		func(ctx context.Context, UUID string) error {
			return api.DeleteAccessReusablePolicy(ctx, UUID)
		},
	)

	//
	tryDeleteAllResourcesFromStore(
		"Removing all created access groups in this test set",
		&tracer.AccessGroups,
		&errs,
		func(ctx context.Context, UUID string) error {
			return api.DeleteAccessGroup(ctx, UUID)
		},
	)

	//
	tryDeleteAllResourcesFromStore(
		"Removing all created service tokens in this test set",
		&tracer.AccessServiceTokens,
		&errs,
		func(ctx context.Context, UUID string) error {
			return api.DeleteAccessServiceToken(ctx, UUID)
		},
	)

	return errs
}

func tryDeleteAllResourcesFromStore(
	descr string,
	store *CfUUIDStore,
	errs *[]error,
	deleteFunc func(ctx context.Context, UUID string) error,
) {
	if len(*store) == 0 {
		return
	}

	ctx := context.Background()

	gin.By(descr)

	//
	for UUID, descr := range *store {
		err := deleteFunc(ctx, UUID)
		if err != nil {
			*errs = append(*errs,
				fault.Wrap(err,
					fctx.With(ctx,
						"description", descr,
					),
				),
			)
		}
	}
}
