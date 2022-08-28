package logblobfinder

import (
	"fmt"
	"testing"
	"time"

	"github.com/tmeadon/nsgpeek/internal/nsgpeektest"
	"github.com/tmeadon/nsgpeek/pkg/azure"
)

func TestFindLatest(t *testing.T) {
	// fakeNsgName := "nsg-view"
	// fakeBlobUrl := "https://path.to/blob"
	// fakeBlob := new(azure.Blob)
	// blobCh := make(chan (*azure.Blob))
	// errCh := make(chan (error))
	// mockStorageBlobGetter := nsgpeektest.NewFakeStorageBlobGetter(fakeBlob, []string{
	// 	fmt.Sprintf("/subscriptions/xxxx/resourceGroups/xxxx/providers/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=05/d=01/h=12/m=00/macAddress=abc/PT1H.json", fakeNsgName),
	// 	fmt.Sprintf("/SUBSCRIPTIONS/xxxx/RESOURCEGROUPS/xxxx/PROVIDERS/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=05/d=01/h=12/m=00/macAddress=abc/PT1H.json", "blah"),
	// 	"abc",
	// 	"asdji2wd29jasdjio2/kla0/!?!(*",
	// 	"123/\\///",
	// })
	var fakeNsgName string
	var fakeBlobUrl string
	var fakeBlobs []*azure.Blob
	var blobCh chan *azure.Blob
	var errCh chan error
	var prefixCh chan string
	var mockStorageBlobGetter *nsgpeektest.StorageBlobGetter
	var finder Finder

	setup := func() {
		fakeNsgName = "nsg-view"
		fakeBlobUrl = "https://path.to/blob"
		fakeBlobs = []*azure.Blob{{Path: "0"}, {Path: "1"}, {Path: "2"}}
		blobCh = make(chan (*azure.Blob))
		errCh = make(chan (error))
		prefixCh = make(chan (string), 5)
		mockStorageBlobGetter = nsgpeektest.NewFakeStorageBlobGetter(fakeBlobs, []string{
			fmt.Sprintf("/subscriptions/xxxx/resourceGroups/xxxx/providers/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=05/d=01/h=12/m=00/macAddress=abc/PT1H.json", fakeNsgName),
			fmt.Sprintf("/SUBSCRIPTIONS/xxxx/RESOURCEGROUPS/xxxx/PROVIDERS/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=05/d=01/h=12/m=00/macAddress=abc/PT1H.json", "blah"),
			"abc",
			"asdji2wd29jasdjio2/kla0/!?!(*",
			"123/\\///",
		}, prefixCh)
		finder = Finder{mockStorageBlobGetter, fakeNsgName}
		overrideGetBlobUrl(fakeBlobUrl)
	}

	// finder := Finder{
	// 	storageBlobGetter: mockStorageBlobGetter,
	// 	nsgName:           fakeNsgName,
	// }

	t.Run("TestFindLatestGetsNewestBlobWithCorrectPrefix", func(t *testing.T) {
		setup()
		expectedPrefix := fmt.Sprintf("/subscriptions/xxxx/resourceGroups/xxxx/providers/microsoft.network/NETWORKSECURITYGROUPS/%v/", fakeNsgName)
		go finder.FindLatest(blobCh, errCh, time.Second*3)
		searchedPrefix := <-prefixCh

		waitForBlob(t, blobCh, errCh, fakeBlobs[0], time.Second*5)

		if searchedPrefix != expectedPrefix {
			t.Errorf("incorrect blob prefix searched.  expected: '%v', got '%v'", expectedPrefix, searchedPrefix)
		}
	})

	t.Run("TestFindLatestSendsNewBlob", func(t *testing.T) {
		setup()
		go finder.FindLatest(blobCh, errCh, time.Second*2)

		waitForBlob(t, blobCh, errCh, fakeBlobs[0], time.Second*5)

		// change the newest blob
		overrideGetBlobUrl(fakeBlobUrl + "/new")
		waitForBlob(t, blobCh, errCh, fakeBlobs[1], time.Second*5)
	})
}

func overrideGetBlobUrl(url string) {
	getBlobUrl = func(*azure.Blob) string {
		return url
	}
}

func waitForBlob(t *testing.T, blobCh chan (*azure.Blob), errCh chan (error), expectedBlob *azure.Blob, timeout time.Duration) {
	select {
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
