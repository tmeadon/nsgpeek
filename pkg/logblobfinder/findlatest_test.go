package logblobfinder

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tmeadon/nsgpeek/internal/nsgpeektest"
	"github.com/tmeadon/nsgpeek/pkg/azure"
)

func TestFindLatest(t *testing.T) {
	fakeNsgName := "nsg-view"
	incorrectNsgName := "blah"
	fakeBlobUrl := "https://path.to/blob"
	fakeBlobs := []*azure.Blob{{Path: "0"}, {Path: "1"}, {Path: "2"}}

	var blobCh chan *azure.Blob
	var errCh chan error
	var prefixCh chan string
	var mockStorageBlobGetter *nsgpeektest.StorageBlobGetter
	var finder Finder

	overrideGetBlobUrl := func(url string) {
		getBlobUrl = func(*azure.Blob) string {
			return url
		}
	}

	setup := func() {
		blobCh = make(chan (*azure.Blob), 5)
		errCh = make(chan (error))
		prefixCh = make(chan (string))
		mockStorageBlobGetter = nsgpeektest.NewFakeStorageBlobGetter(fakeBlobs, []azure.Blob{
			{
				Path:         fmt.Sprintf("/SUBSCRIPTIONS/xxxx/RESOURCEGROUPS/xxxx/PROVIDERS/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=05/d=01/h=12/m=00/macAddress=abc/PT1H.json", incorrectNsgName),
				LastModified: time.Date(2022, 5, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				Path:         fmt.Sprintf("/SUBSCRIPTIONS/xxxx/RESOURCEGROUPS/xxxx/PROVIDERS/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=03/d=01/h=12/m=00/macAddress=abc/PT1H.json", incorrectNsgName),
				LastModified: time.Date(2022, 3, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				Path:         fmt.Sprintf("/subscriptions/xxxx/resourceGroups/xxxx/providers/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=05/d=01/h=12/m=00/macAddress=abc/PT1H.json", fakeNsgName),
				LastModified: time.Date(2022, 5, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				Path:         fmt.Sprintf("/subscriptions/xxxx/resourceGroups/xxxx/providers/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=04/d=01/h=12/m=00/macAddress=abc/PT1H.json", fakeNsgName),
				LastModified: time.Date(2022, 5, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				Path:         fmt.Sprintf("/subscriptions/xxxx/resourceGroups/xxxx/providers/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2021/m=12/d=01/h=12/m=00/macAddress=abc/PT1H.json", fakeNsgName),
				LastModified: time.Date(2022, 5, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				Path:         "abc",
				LastModified: time.Date(2022, 5, 1, 1, 0, 0, 0, time.UTC),
			},
			{
				Path:         "asdji2wd29jasdjio2/kla0/!?!(*",
				LastModified: time.Date(2022, 5, 1, 1, 0, 0, 0, time.UTC),
			},
		}, prefixCh, true)
		finder = Finder{mockStorageBlobGetter, fakeNsgName}
		overrideGetBlobUrl(fakeBlobUrl)
	}

	waitForBlob := func(t *testing.T, expectedBlob *azure.Blob, timeout time.Duration) {
		select {
		case <-prefixCh:

		case blob := <-blobCh:
			if blob != expectedBlob {
				t.Errorf("wrong blob received from goroutine. expected: %v; got: %v", expectedBlob, blob)
			}

		case err := <-errCh:
			t.Errorf("unexpected error received: %v", err)

		case <-time.After(timeout):
			t.Errorf("timed out waiting for latest blob to be found")
		}
	}

	t.Run("SearchesForBlobWithCorrectPrefix", func(t *testing.T) {
		setup()
		go finder.FindLatest(blobCh, errCh, time.Second*3)
		searchedPrefixes := make([]string, 0)

	wait:
		for {
			select {
			case p := <-prefixCh:
				searchedPrefixes = append(searchedPrefixes, p)
			case <-blobCh:
				break wait
			}
		}

		// check that the prefix for the incorrect nsg name wasn't searched
		for _, p := range searchedPrefixes {
			if strings.Contains(p, incorrectNsgName) {
				t.Errorf("incorrect blob prefix searched: %v", p)
			}
		}
	})

	t.Run("SendsNewBlob", func(t *testing.T) {
		setup()
		go finder.FindLatest(blobCh, errCh, time.Second*2)

		waitForBlob(t, fakeBlobs[0], time.Second*5)

		// change the newest blob
		overrideGetBlobUrl(fakeBlobUrl + "/new")
		waitForBlob(t, fakeBlobs[1], time.Second*5)
	})
}
