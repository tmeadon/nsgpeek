package logblobfinder

import (
	"sort"
	"time"

	"github.com/tmeadon/nsgpeek/pkg/azure"
)

var getBlobUrl = func(blob *azure.Blob) string {
	return blob.URL()
}

func (f *Finder) FindLatest(ch chan (*azure.Blob), errCh chan (error), sleepDuration time.Duration) {
	logPrefix, err := f.findNsgBlobPrefix()
	if err != nil {
		errCh <- err
		return
	}

	var currentBlobUrl string

	for {
		newestBlob, err := f.findNewestBlob(logPrefix)

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

func (f *Finder) findNewestBlob(prefix string) (*azure.Blob, error) {
	blobs, childPrefixes, err := f.ListBlobDirectory(prefix)
	if err != nil {
		return nil, err
	}

	if len(childPrefixes) == 0 {
		var newestBlob azure.Blob

		for _, b := range blobs {
			if b.LastModified.After(newestBlob.LastModified) {
				newestBlob = b
			}
		}

		return &newestBlob, nil
	}

	newestPrefix := getNewestPrefix(childPrefixes)
	newestBlob, err := f.findNewestBlob(newestPrefix)
	return newestBlob, err
}

func getNewestPrefix(prefixes []string) string {
	sort.Strings(prefixes)
	return prefixes[len(prefixes)-1]
}
