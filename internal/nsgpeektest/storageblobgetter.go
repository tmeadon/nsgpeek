package nsgpeektest

import (
	"fmt"
	"regexp"

	"github.com/tmeadon/nsgpeek/pkg/azure"
)

type StorageBlobGetter struct {
	newestBlobs     []*azure.Blob
	newestBlobIndex int
	PrefixChan      chan string
	Blobs           []string
}

func NewFakeStorageBlobGetter(newestBlobs []*azure.Blob, fakeBlobPaths []string, prefixCh chan string) *StorageBlobGetter {
	return &StorageBlobGetter{
		newestBlobs:     newestBlobs,
		newestBlobIndex: 0,
		Blobs:           fakeBlobPaths,
		PrefixChan:      prefixCh,
	}
}

func (f *StorageBlobGetter) GetNewestBlob(prefix string) (*azure.Blob, error) {
	f.PrefixChan <- prefix
	blob := f.newestBlobs[f.newestBlobIndex]
	f.newestBlobIndex++
	return blob, nil
}

func (f *StorageBlobGetter) ListBlobDirectory(prefix string) (blobs []string, prefixes []string, err error) {
	re := f.blobPrefixRegex(prefix)
	prefixMap := make(map[string]bool)
	blobMap := make(map[string]bool)

	for _, path := range f.Blobs {
		matches := re.FindStringSubmatch(path)

		if len(matches) > 1 {
			m := matches[1]

			if m[len(m)-1] == '/' {
				prefixMap[m] = true
			} else {
				blobMap[m] = true
			}
		}
	}

	for k := range prefixMap {
		prefixes = append(prefixes, k)
	}

	for k := range blobMap {
		blobs = append(blobs, k)
	}

	return
}

func (f *StorageBlobGetter) ListBlobs(prefix string) (blobs []azure.Blob, err error) {
	re := f.blobPrefixRegex(prefix)

	for _, path := range f.Blobs {
		if re.Match([]byte(path)) {
			blobs = append(blobs, azure.Blob{Path: path})
		}
	}

	return
}

func (f *StorageBlobGetter) blobPrefixRegex(prefix string) *regexp.Regexp {
	pattern := fmt.Sprintf(`(?i)^(%v[^\/]*[\/]?).*$`, regexp.QuoteMeta(prefix))
	re := regexp.MustCompile(pattern)
	return re
}
