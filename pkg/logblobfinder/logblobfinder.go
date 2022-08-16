package logblobfinder

import (
	"context"

	"github.com/tmeadon/nsgpeek/pkg/azure"
)

type nsgGetter interface {
	GetNsgFlowLogStorageId(subscriptionIds []string) (*azure.ResourceId, error)
}

type storageBlobGetter interface {
	GetNewestBlob(stgAccId *azure.ResourceId) (*azure.Blob, error)
}

type Finder struct {
	nsgGetter
	storageBlobGetter
	allSubscriptionIds []string
}

func NewLogBlobFinder(allSubscriptionIds []string, nsgName string, ctx context.Context, cred *azure.Credential) *Finder {
	nsgGetter := azure.NewAzureNsgGetter(nsgName, ctx, cred)
	blobGetter := azure.NewAzureStorageBlobGetter(ctx, cred)

	return &Finder{
		nsgGetter:          nsgGetter,
		storageBlobGetter:  blobGetter,
		allSubscriptionIds: allSubscriptionIds,
	}
}
