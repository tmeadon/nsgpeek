package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var (
	subscriptionID           string = ""
	nsgRg                    string = "nsg-view"
	nsgName                  string = "nsg-view"
	flowLogBlobContainerName string = "insights-logs-networksecuritygroupflowevent"
)

func main() {
	// subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set")
	}

	// auth
	log.Print("authenticating")

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	ctx := context.Background()

	// get nsg flow log details
	log.Print("create nsg client")

	nsgClient, err := armnetwork.NewSecurityGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("get nsg")

	expand := "flowLogs"
	nsg, err := nsgClient.Get(ctx, nsgRg, nsgName, &armnetwork.SecurityGroupsClientGetOptions{Expand: &expand})
	if err != nil {
		log.Fatal(err)
	}

	logStgID, err := arm.ParseResourceID(*nsg.Properties.FlowLogs[0].Properties.StorageID)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("flow logs storage ID: %v", logStgID.String())

	// get storage access key
	stgClient, err := armstorage.NewAccountsClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	keys, err := stgClient.ListKeys(ctx, logStgID.ResourceGroupName, logStgID.Name, nil)
	if err != nil {
		log.Fatal(err)
	}

	// find the most recent flow log blob
	blobCred, err := azblob.NewSharedKeyCredential(logStgID.Name, *keys.Keys[0].Value)
	if err != nil {
		log.Fatal(err)
	}

	serviceClient, err := azblob.NewServiceClientWithSharedKey(fmt.Sprintf("https://%s.blob.core.windows.net/", logStgID.Name), blobCred, nil)
	if err != nil {
		log.Fatal(err)
	}

	containerClient, err := serviceClient.NewContainerClient(flowLogBlobContainerName)
	if err != nil {
		log.Fatal(err)
	}

	pager := containerClient.ListBlobsFlat(nil)
	var newestBlob *azblob.BlobItemInternal

	for pager.NextPage(ctx) {
		resp := pager.PageResponse()

		for _, v := range resp.ListBlobsFlatSegmentResponse.Segment.BlobItems {
			if newestBlob == nil || newestBlob.Properties.LastModified.Before(*v.Properties.LastModified) {
				newestBlob = v
			}
		}
	}

	if err = pager.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("newest blob: %v, last modified: %v", *newestBlob.Name, *newestBlob.Properties.LastModified)

	// create a blob clients and retrieve the block list
	blob, err := containerClient.NewBlockBlobClient(*newestBlob.Name)
	if err != nil {
		log.Fatal(err)
	}

	blocks, err := blob.GetBlockList(ctx, azblob.BlockListTypeAll, nil)
	if err != nil {
		log.Fatal(err)
	}

	// get the size of the biggest block
	biggest := 0

	for _, b := range blocks.CommittedBlocks {
		if int(*b.Size) > biggest {
			biggest = int(*b.Size)
		}
	}

	// read each block one by one
	index := int64(0)
	values := make([]string, 0)

	for _, block := range blocks.CommittedBlocks {
		blockGet, err := blob.Download(ctx, &azblob.BlobDownloadOptions{
			Count:  block.Size,
			Offset: &index,
		})
		if err != nil {
			log.Fatal(err)
		}

		data := &bytes.Buffer{}
		reader := blockGet.Body(&azblob.RetryReaderOptions{})
		_, err = data.ReadFrom(reader)
		if err != nil {
			log.Fatal(err)
		}

		err = reader.Close()
		if err != nil {
			log.Fatal(err)
		}

		index = index + *block.Size
		values = append(values, data.String())
	}

	for _, v := range values {
		fmt.Printf("block contents: %v\n\n", v)
	}
}
