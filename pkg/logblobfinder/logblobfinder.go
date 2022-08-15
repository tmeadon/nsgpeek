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

type LogBlobFinder struct {
	nsgGetter
	storageBlobGetter
	allSubscriptionIds []string
}

func NewLogBlobFinder(allSubscriptionIds []string, nsgName string, ctx context.Context, cred *azure.Credential) *LogBlobFinder {
	nsgGetter := azure.NewAzureNsgGetter(nsgName, ctx, cred)
	blobGetter := azure.NewAzureStorageBlobGetter(ctx, cred)

	return &LogBlobFinder{
		nsgGetter:          nsgGetter,
		storageBlobGetter:  blobGetter,
		allSubscriptionIds: allSubscriptionIds,
	}
}
