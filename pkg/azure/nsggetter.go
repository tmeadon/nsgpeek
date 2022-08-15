package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

type AzureNsgGetter struct {
	nsgName string
	ctx     context.Context
	cred    *Credential
}

func NewAzureNsgGetter(nsgName string, ctx context.Context, cred *Credential) *AzureNsgGetter {
	return &AzureNsgGetter{
		nsgName: nsgName,
		ctx:     ctx,
		cred:    cred,
	}
}

func (a *AzureNsgGetter) GetNsgFlowLogStorageId(subscriptionIds []string) (*ResourceId, error) {
	nsgId, err := a.findNsg(subscriptionIds)
	if err != nil {
		return nil, err
	}

	nsg, err := a.getNsgById(nsgId)
	if err != nil {
		return nil, err
	}

	stgId, err := arm.ParseResourceID(*nsg.Properties.FlowLogs[0].Properties.StorageID)
	if err != nil {
		return nil, err
	}

	return &ResourceId{*stgId}, nil
}

func (a *AzureNsgGetter) newNsgClient(subscriptionId string) (*armnetwork.SecurityGroupsClient, error) {
	c, err := armnetwork.NewSecurityGroupsClient(subscriptionId, *a.cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create nsg client: %w", err)
	}
	return c, nil
}

func (a *AzureNsgGetter) findNsg(subscriptionIds []string) (*arm.ResourceID, error) {
	for _, subId := range subscriptionIds {
		nsgId, err := a.searchSubForNsg(subId, a.nsgName)
		if err != nil {
			return nil, err
		}
		if nsgId != nil {
			return nsgId, nil
		}
	}

	return nil, fmt.Errorf("could not find nsg '%v' in subscriptions: %v", a.nsgName, subscriptionIds)
}

func (a *AzureNsgGetter) searchSubForNsg(subscriptionId string, nsgName string) (*arm.ResourceID, error) {
	client, err := a.newNsgClient(subscriptionId)
	if err != nil {
		return nil, err
	}

	pager := client.NewListAllPager(nil)

	for pager.More() {
		page, err := pager.NextPage(a.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve nsgs: %w", err)
		}

		for _, n := range page.Value {
			if *n.Name == nsgName {
				r, err := arm.ParseResourceID(*n.ID)
				if err != nil {
					return nil, fmt.Errorf("could not parse nsg resource id %v: %w", n.ID, err)
				}

				return r, nil
			}
		}
	}

	return nil, nil
}

func (a *AzureNsgGetter) getNsgById(nsgId *arm.ResourceID) (*armnetwork.SecurityGroupsClientGetResponse, error) {
	client, err := a.newNsgClient(nsgId.SubscriptionID)
	if err != nil {
		return nil, err
	}

	expand := "flowLogs"
	nsg, err := client.Get(a.ctx, nsgId.ResourceGroupName, nsgId.Name, &armnetwork.SecurityGroupsClientGetOptions{Expand: &expand})
	if err != nil {
		return nil, fmt.Errorf("failed to get nsg: %w", err)
	}

	return &nsg, nil
}
