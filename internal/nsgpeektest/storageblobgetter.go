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
	SendPrefixes    bool
	Blobs           []azure.Blob
}

func NewFakeStorageBlobGetter(newestBlobs []*azure.Blob, fakeBlobs []azure.Blob, prefixCh chan string, sendPrefix bool) *StorageBlobGetter {
	return &StorageBlobGetter{
		newestBlobs:     newestBlobs,
		newestBlobIndex: 0,
		Blobs:           fakeBlobs,
		PrefixChan:      prefixCh,
		SendPrefixes:    sendPrefix,
	}
}

func (f *StorageBlobGetter) ListBlobDirectory(prefix string) (blobs []azure.Blob, prefixes []string, err error) {
	f.sendPrefix(prefix)
	re := f.blobPrefixRegex(prefix)
	prefixMap := make(map[string]bool)
	blobMap := make(map[string]bool)

	for _, blob := range f.Blobs {
		matches := re.FindStringSubmatch(blob.Path)

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
		for _, blob := range f.Blobs {
			if blob.Path == k {
				blobs = append(blobs, blob)
			}
		}
	}

	return
}

func (f *StorageBlobGetter) ListBlobs(prefix string) (blobs []azure.Blob, err error) {
	re := f.blobPrefixRegex(prefix)

	for _, blob := range f.Blobs {
		if re.Match([]byte(blob.Path)) {
			blobs = append(blobs, blob)
		}
	}

	return
}

func (f *StorageBlobGetter) blobPrefixRegex(prefix string) *regexp.Regexp {
	pattern := fmt.Sprintf(`(?i)^(%v[^\/]*[\/]?).*$`, regexp.QuoteMeta(prefix))
	re := regexp.MustCompile(pattern)
	return re
}

func (f *StorageBlobGetter) sendPrefix(prefix string) {
	if f.SendPrefixes {
		f.PrefixChan <- prefix
	}
}
