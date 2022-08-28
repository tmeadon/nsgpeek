package nsgpeektest

import (
	"fmt"
	"regexp"

	"github.com/tmeadon/nsgpeek/pkg/azure"
)

type StorageBlobGetter struct {
	NewestBlob             *azure.Blob
	NewestBlobSearchPrefix string
	Blobs                  []string
}

func NewFakeStorageBlobGetter(newestBlob *azure.Blob, fakeBlobPaths []string) *StorageBlobGetter {
	return &StorageBlobGetter{
		NewestBlob: newestBlob,
		Blobs:      fakeBlobPaths,
	}
}

func (f *StorageBlobGetter) GetNewestBlob(prefix string) (*azure.Blob, error) {
	f.NewestBlobSearchPrefix = prefix
	return f.NewestBlob, nil
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
