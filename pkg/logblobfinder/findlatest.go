package logblobfinder

import (
	"fmt"
	"time"

	"github.com/tmeadon/nsgpeek/pkg/azure"
)

var getBlobUrl = func(blob *azure.Blob) string {
	return blob.URL()
}

func (f *LogBlobFinder) FindLatest(ch chan (*azure.Blob), errCh chan (error), sleepDuration time.Duration) {
	stgId, err := f.GetNsgFlowLogStorageId(f.allSubscriptionIds)
	if err != nil {
		errCh <- fmt.Errorf("unable to find storage id for flow logs: %w", err)
		return
	}

	var currentBlobUrl string

	for {
		newestBlob, err := f.GetNewestBlob(stgId)
		if err != nil {
			errCh <- err
			return
		}

		blobUrl := getBlobUrl(newestBlob)

		if currentBlobUrl != blobUrl {
			ch <- newestBlob
			currentBlobUrl = blobUrl
		}

		time.Sleep(sleepDuration)
	}
}
