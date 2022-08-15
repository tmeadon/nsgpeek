package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var flowLogBlobContainerName string = "insights-logs-networksecuritygroupflowevent"

type Blob struct {
	azblob.BlockBlobClient
}

type AzureStorageBlobGetter struct {
	ctx  context.Context
	cred *azcore.TokenCredential
}

func NewAzureStorageBlobGetter(ctx context.Context, cred *azcore.TokenCredential) *AzureStorageBlobGetter {
	return &AzureStorageBlobGetter{
		ctx:  ctx,
		cred: cred,
	}
}

func (a *AzureStorageBlobGetter) GetNewestBlob(stgAccId *ResourceId) (*Blob, error) {
	client, err := a.getContainerClient(stgAccId)
	if err != nil {
		return nil, err
	}

	newestBlobName, err := a.findNewestBlobName(client)
	if err != nil {
		return nil, err
	}

	blobClient, err := a.getBlockBlobClient(client, newestBlobName)
	if err != nil {
		return nil, err
	}

	return &Blob{*blobClient}, nil
}

func (a *AzureStorageBlobGetter) getContainerClient(stgAccId *ResourceId) (*azblob.ContainerClient, error) {
	stgAccClient, err := a.newStorageAccountClient(stgAccId.SubscriptionID)
	if err != nil {
		return nil, err
	}

	keys, err := stgAccClient.ListKeys(a.ctx, stgAccId.ResourceGroupName, stgAccId.Name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage keys: %w", err)
	}

	blobCred, err := azblob.NewSharedKeyCredential(stgAccId.Name, *keys.Keys[0].Value)
	if err != nil {
		return nil, fmt.Errorf("failed to created blob credential: %w", err)
	}

	serviceClient, err := azblob.NewServiceClientWithSharedKey(fmt.Sprintf("https://%s.blob.core.windows.net/", stgAccId.Name), blobCred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob service client: %w", err)
	}

	containerClient, err := serviceClient.NewContainerClient(flowLogBlobContainerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container client: %w", err)
	}

	return containerClient, nil
}

func (a *AzureStorageBlobGetter) newStorageAccountClient(subscriptionId string) (*armstorage.AccountsClient, error) {
	stgClient, err := armstorage.NewAccountsClient(subscriptionId, *a.cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage account client: %w", err)
	}
	return stgClient, nil
}

func (a *AzureStorageBlobGetter) findNewestBlobName(containerClient *azblob.ContainerClient) (string, error) {
	var newestBlob *azblob.BlobItemInternal
	pager := containerClient.ListBlobsFlat(nil)

	for pager.NextPage(a.ctx) {
		resp := pager.PageResponse()

		for _, v := range resp.ListBlobsFlatSegmentResponse.Segment.BlobItems {
			if newestBlob == nil || newestBlob.Properties.LastModified.Before(*v.Properties.LastModified) {
				newestBlob = v
			}
		}
	}

	if err := pager.Err(); err != nil {
		return "", fmt.Errorf("failed to list blobs: %w", err)
	}

	return *newestBlob.Name, nil
}

func (a *AzureStorageBlobGetter) getBlockBlobClient(containerClient *azblob.ContainerClient, blobName string) (*azblob.BlockBlobClient, error) {
	blob, err := containerClient.NewBlockBlobClient(blobName)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob client: %w", err)
	}

	return blob, nil
}
