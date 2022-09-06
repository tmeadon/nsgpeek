package logblobfinder

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/tmeadon/nsgpeek/internal/nsgpeektest"
	"github.com/tmeadon/nsgpeek/pkg/azure"
)

func TestFindSpecific(t *testing.T) {
	fakeNsgName := "nsg-view"
	fakeBlob := new(azure.Blob)

	t.Run("GetsTheCorrectBlobs", func(t *testing.T) {
		start := time.Date(2022, 01, 01, 0, 0, 0, 0, time.UTC)
		end := time.Date(2022, 01, 02, 0, 0, 0, 0, time.UTC)
		goodBlobs := []azure.Blob{
			{
				Path:         fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=13/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				LastModified: time.Date(2022, 1, 1, 13, 59, 0, 0, time.UTC),
			},
			{
				Path:         fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=19/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				LastModified: time.Date(2022, 1, 1, 19, 52, 0, 0, time.UTC),
			},
			{
				Path:         fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=00/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				LastModified: time.Date(2022, 1, 2, 0, 58, 0, 0, time.UTC),
			},
		}
		badBlobs := []azure.Blob{
			{
				Path:         fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=02/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				LastModified: time.Date(2022, 1, 2, 2, 58, 0, 0, time.UTC),
			},
			{
				Path:         fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=02/m=00/macAddress=0022483F762A/PT1H.json", "bad"),
				LastModified: time.Date(2022, 1, 1, 2, 58, 0, 0, time.UTC),
			},
		}

		blobGetter := nsgpeektest.NewFakeStorageBlobGetter([]*azure.Blob{fakeBlob}, append(goodBlobs, badBlobs...), make(chan string), false)
		finder := Finder{blobGetter, fakeNsgName}

		got, err := finder.FindSpecific(start, end)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !nsgpeektest.BlobSlicesEqual(got, goodBlobs) {
			t.Fatalf("incorect blobs returned, expected: %#v, got: %#v", goodBlobs, got)
		}
	})

	t.Run("ReturnsNothingIfDatesReversed", func(t *testing.T) {
		end := time.Date(2022, 01, 01, 0, 0, 0, 0, time.UTC)
		start := time.Date(2022, 01, 02, 0, 0, 0, 0, time.UTC)
		blobs := []azure.Blob{
			{
				Path:         fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=13/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				LastModified: time.Date(2022, 1, 1, 13, 59, 0, 0, time.UTC),
			},
			{
				Path:         fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=00/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				LastModified: time.Date(2022, 1, 2, 0, 58, 0, 0, time.UTC),
			},
			{
				Path:         fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=02/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				LastModified: time.Date(2022, 1, 2, 2, 58, 0, 0, time.UTC),
			},
			{
				Path:         fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=02/m=00/macAddress=0022483F762A/PT1H.json", "bad"),
				LastModified: time.Date(2022, 1, 1, 2, 58, 0, 0, time.UTC),
			},
		}

		blobGetter := nsgpeektest.NewFakeStorageBlobGetter([]*azure.Blob{fakeBlob}, blobs, make(chan string), false)
		finder := Finder{blobGetter, fakeNsgName}

		got, err := finder.FindSpecific(start, end)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) > 0 {
			t.Fatalf("expected to receive no blobs, instead received %#v", got)
		}
	})

	t.Run("ReturnsCorrectErrorIfNoBlobsForNsg", func(t *testing.T) {
		start := time.Date(2022, 01, 01, 0, 0, 0, 0, time.UTC)
		end := time.Date(2022, 01, 02, 0, 0, 0, 0, time.UTC)
		blobs := []azure.Blob{
			{
				Path:         fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=02/m=00/macAddress=0022483F762A/PT1H.json", "bad"),
				LastModified: time.Date(2022, 1, 1, 2, 58, 0, 0, time.UTC),
			},
		}

		blobGetter := nsgpeektest.NewFakeStorageBlobGetter([]*azure.Blob{fakeBlob}, blobs, make(chan string), false)
		finder := Finder{blobGetter, fakeNsgName}

		_, err := finder.FindSpecific(start, end)

		if !errors.Is(err, ErrBlobPrefixNotFound) {
			t.Fatalf("expected error: %v, got: %v", ErrBlobPrefixNotFound, err)
		}
	})
}
