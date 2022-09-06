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
	var fakeNsgName string
	var fakeBlobUrl string
	var fakeBlobs []*azure.Blob
	var blobCh chan *azure.Blob
	var errCh chan error
	var prefixCh chan string
	var mockStorageBlobGetter *nsgpeektest.StorageBlobGetter
	var finder Finder
	var incorrectNsgName string

	setup := func() {
		fakeNsgName = "nsg-view"
		incorrectNsgName = "blah"
		fakeBlobUrl = "https://path.to/blob"
		fakeBlobs = []*azure.Blob{{Path: "0"}, {Path: "1"}, {Path: "2"}}
		blobCh = make(chan (*azure.Blob), 5)
		errCh = make(chan (error))
		prefixCh = make(chan (string))
		mockStorageBlobGetter = nsgpeektest.NewFakeStorageBlobGetter(fakeBlobs, []string{
			fmt.Sprintf("/SUBSCRIPTIONS/xxxx/RESOURCEGROUPS/xxxx/PROVIDERS/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=05/d=01/h=12/m=00/macAddress=abc/PT1H.json", incorrectNsgName),
			fmt.Sprintf("/SUBSCRIPTIONS/xxxx/RESOURCEGROUPS/xxxx/PROVIDERS/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=03/d=01/h=12/m=00/macAddress=abc/PT1H.json", incorrectNsgName),
			fmt.Sprintf("/subscriptions/xxxx/resourceGroups/xxxx/providers/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=05/d=01/h=12/m=00/macAddress=abc/PT1H.json", fakeNsgName),
			"abc",
			"asdji2wd29jasdjio2/kla0/!?!(*",
			// "123/\\///",
		}, prefixCh)
		finder = Finder{mockStorageBlobGetter, fakeNsgName}
		overrideGetBlobUrl(fakeBlobUrl)
	}

	t.Run("SearchesForBlobWithCorrectPrefix", func(t *testing.T) {
		setup()
		go finder.FindLatest(blobCh, errCh, time.Second*3)
		searchedPrefixes := make([]string, 0)
		expectedPrefix := fmt.Sprintf("/subscriptions/xxxx/resourceGroups/xxxx/providers/microsoft.network/NETWORKSECURITYGROUPS/%v/", fakeNsgName)

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

		// the last element in searchedPrefixes will be the prefix searched in call to GetNewestBlob
		if searchedPrefixes[len(searchedPrefixes)-1] != expectedPrefix {
			t.Errorf("incorrect blob prefix searched.  expected: '%v', got '%v'", expectedPrefix, searchedPrefixes[len(searchedPrefixes)-1])
		}
	})

	t.Run("SendsNewBlob", func(t *testing.T) {
		setup()
		go finder.FindLatest(blobCh, errCh, time.Second*2)

		waitForBlob(t, blobCh, errCh, prefixCh, fakeBlobs[0], time.Second*5)

		// change the newest blob
		overrideGetBlobUrl(fakeBlobUrl + "/new")
		waitForBlob(t, blobCh, errCh, prefixCh, fakeBlobs[1], time.Second*5)
	})
}

func overrideGetBlobUrl(url string) {
	getBlobUrl = func(*azure.Blob) string {
		return url
	}
}

func waitForBlob(t *testing.T, blobCh chan (*azure.Blob), errCh chan (error), prefixCh chan (string), expectedBlob *azure.Blob, timeout time.Duration) {
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
