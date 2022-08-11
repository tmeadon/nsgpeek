package logblobfinder

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var flowLogBlobContainerName string = "insights-logs-networksecuritygroupflowevent"

type LogBlobFinder struct {
	allSubscriptionIds []string
	nsgName            string
	cred               azcore.TokenCredential
	log                *log.Logger
	ctx                context.Context
	flowLogStorageId   *arm.ResourceID
	stgAccClient       *armstorage.AccountsClient
}

func NewLogBlobFinder(allSubscriptionIds []string, nsgName string, log *log.Logger, ctx context.Context, cred azcore.TokenCredential) *LogBlobFinder {
	return &LogBlobFinder{
		allSubscriptionIds: allSubscriptionIds,
		nsgName:            nsgName,
		cred:               cred,
		log:                log,
		ctx:                ctx,
	}
}

func (f *LogBlobFinder) findNsgLogStorageId() error {
	nsgId, err := f.findNsg()
	if err != nil {
		return err
	}

	nsgClient, err := f.nsgNewClient(nsgId.SubscriptionID)
	if err != nil {
		return err
	}

	expand := "flowLogs"
	nsg, err := nsgClient.Get(f.ctx, nsgId.ResourceGroupName, nsgId.Name, &armnetwork.SecurityGroupsClientGetOptions{Expand: &expand})
	if err != nil {
		return fmt.Errorf("failed to retrieve nsg: %w", err)
	}

	log.Printf("found nsg log storage account: %v", *nsg.Properties.FlowLogs[0].Properties.StorageID)

	logStgID, err := arm.ParseResourceID(*nsg.Properties.FlowLogs[0].Properties.StorageID)
	if err != nil {
		return fmt.Errorf("failed to parse storage account ID %v: %w", *nsg.Properties.FlowLogs[0].Properties.StorageID, err)
	}

	f.flowLogStorageId = logStgID
	return nil
}

func (f *LogBlobFinder) findNsg() (*arm.ResourceID, error) {
	for _, subId := range f.allSubscriptionIds {

		log.Printf("looking for nsg in sub: %v", subId)

		nsgClient, err := f.nsgNewClient(subId)
		if err != nil {
			return nil, err
		}

		pager := nsgClient.NewListAllPager(nil)

		for pager.More() {
			page, err := pager.NextPage(f.ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve nsgs: %w", err)
			}

			for _, n := range page.Value {
				if *n.Name == f.nsgName {
					r, err := arm.ParseResourceID(*n.ID)
					if err != nil {
						return nil, fmt.Errorf("could not parse nsg resource id %v: %w", n.ID, err)
					}

					log.Printf("found nsg: %v", r.String())

					return r, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("nsg %v does not exist in subscriptions %v", f.nsgName, f.allSubscriptionIds)
}

func (f *LogBlobFinder) getContainerClient(logStgId *arm.ResourceID) (*azblob.ContainerClient, error) {
	stgAccClient, err := f.newStorageAccountClient(logStgId.SubscriptionID)
	if err != nil {
		return nil, err
	}

	keys, err := stgAccClient.ListKeys(f.ctx, logStgId.ResourceGroupName, logStgId.Name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage keys: %w", err)
	}

	blobCred, err := azblob.NewSharedKeyCredential(logStgId.Name, *keys.Keys[0].Value)
	if err != nil {
		return nil, fmt.Errorf("failed to created blob credential: %w", err)
	}

	serviceClient, err := azblob.NewServiceClientWithSharedKey(fmt.Sprintf("https://%s.blob.core.windows.net/", logStgId.Name), blobCred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob service client: %w", err)
	}

	containerClient, err := serviceClient.NewContainerClient(flowLogBlobContainerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container client: %w", err)
	}

	return containerClient, nil
}

func (f *LogBlobFinder) getBlockBlobClient(containerClient *azblob.ContainerClient, blobName string) (*azblob.BlockBlobClient, error) {
	blob, err := containerClient.NewBlockBlobClient(blobName)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob client: %w", err)
	}

	return blob, nil
}

func (f *LogBlobFinder) nsgNewClient(subscriptionId string) (*armnetwork.SecurityGroupsClient, error) {
	nsgClient, err := armnetwork.NewSecurityGroupsClient(subscriptionId, f.cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create nsg client: %w", err)
	}
	return nsgClient, nil
}

func (f *LogBlobFinder) newStorageAccountClient(subscriptionId string) (*armstorage.AccountsClient, error) {
	stgClient, err := armstorage.NewAccountsClient(subscriptionId, f.cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage account client: %w", err)
	}
	return stgClient, nil
}
