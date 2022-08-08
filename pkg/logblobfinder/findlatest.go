package logblobfinder

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func (f *LogBlobFinder) FindLatest(ch chan (*azblob.BlockBlobClient), errCh chan (error)) {
	if f.flowLogStorageId == nil {
		err := f.findNsgLogStorageId()
		if err != nil {
			errCh <- fmt.Errorf("unable to find storage id for flow logs: %w", err)
			return
		}
	}

	containerClient, err := f.getContainerClient(f.flowLogStorageId)
	if err != nil {
		errCh <- err
		return
	}

	var currentBlob string

	for {
		newestBlob, err := f.findNewestBlob(containerClient)
		if err != nil {
			errCh <- err
			return
		}

		if currentBlob != newestBlob {
			blobClient, err := f.getBlockBlobClient(containerClient, newestBlob)
			if err != nil {
				errCh <- err
				return
			}
			ch <- blobClient
			currentBlob = newestBlob
		}

		time.Sleep(time.Second * 10)
	}
}

func (f *LogBlobFinder) findNewestBlob(containerClient *azblob.ContainerClient) (string, error) {
	var newestBlob *azblob.BlobItemInternal
	pager := containerClient.ListBlobsFlat(nil)

	for pager.NextPage(f.ctx) {
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
