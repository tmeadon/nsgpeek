package logblobfinder

import (
	"time"

	"github.com/tmeadon/nsgpeek/pkg/azure"
)

var getBlobUrl = func(blob *azure.Blob) string {
	return blob.URL()
}

func (f *Finder) FindLatest(ch chan (*azure.Blob), errCh chan (error), sleepDuration time.Duration) {
	logPrefix, err := f.FindNsgBlobPrefix()
	if err != nil {
		errCh <- err
		return
	}

	var currentBlobUrl string

	for {
		newestBlob, err := f.GetNewestBlob(logPrefix)
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
