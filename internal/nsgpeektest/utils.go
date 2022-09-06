package nsgpeektest

import (
	"sort"

	"github.com/tmeadon/nsgpeek/pkg/azure"
)

func SortBlobs(blobs []azure.Blob) {
	sort.Slice(blobs, func(i, j int) bool {
		return blobs[i].Path < blobs[j].Path
	})
}

func BlobSlicesEqual(a []azure.Blob, b []azure.Blob) bool {
	SortBlobs(a)
	SortBlobs(b)

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Path != b[i].Path || !a[i].LastModified.Equal(b[i].LastModified) {
			return false
		}
	}

	return true
}
