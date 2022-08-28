package logblobfinder

import (
	"context"
	"fmt"
	"regexp"

	"github.com/tmeadon/nsgpeek/pkg/azure"
)

type storageBlobGetter interface {
	GetNewestBlob(prefix string) (*azure.Blob, error)
	ListBlobDirectory(prefix string) ([]string, []string, error)
	ListBlobs(prefix string) ([]azure.Blob, error)
}

type Finder struct {
	storageBlobGetter
	nsgName string
}

func NewLogBlobFinder(allSubscriptionIds []string, nsgName string, ctx context.Context, cred *azure.Credential) (*Finder, error) {
	nsgGetter := azure.NewAzureNsgGetter(nsgName, ctx, cred)
	stgId, err := nsgGetter.GetNsgFlowLogStorageId(allSubscriptionIds)
	if err != nil {
		return nil, fmt.Errorf("failed to get nsg log storage id: %w", err)
	}

	blobGetter, err := azure.NewAzureStorageBlobGetter(ctx, cred, stgId)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob getter: %w", err)
	}

	return &Finder{
		storageBlobGetter: blobGetter,
		nsgName:           nsgName,
	}, nil
}

func (f *Finder) FindNsgBlobPrefix() (string, error) {
	p, err := f.findBlobPrefix("")
	return p, err
}

func (f *Finder) findBlobPrefix(prefix string) (string, error) {
	_, prefixes, err := f.storageBlobGetter.ListBlobDirectory(prefix)
	if err != nil {
		return "", err
	}

	for _, p := range prefixes {
		if isMatch(p, f.nsgName) {
			return p, nil
		} else {
			found, err := f.findBlobPrefix(p)

			if err != nil {
				return "", err
			}

			if found != "" {
				return found, nil
			}
		}
	}

	return "", nil
}

func isMatch(path string, nsgName string) bool {
	r := regexp.MustCompile(`(?i).*\/networksecuritygroups\/` + nsgName + `\/$`)
	m := r.Match([]byte(path))
	return m
}
