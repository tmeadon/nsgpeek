package logblobfinder

import (
	"fmt"
	"testing"
	"time"

	"github.com/tmeadon/nsgpeek/internal/nsgpeektest"
	"github.com/tmeadon/nsgpeek/pkg/azure"
)

func TestFindLatest(t *testing.T) {
	fakeNsgName := "nsg-view"
	fakeBlobUrl := "https://path.to/blob"
	fakeBlob := new(azure.Blob)
	blobCh := make(chan (*azure.Blob))
	errCh := make(chan (error))
	mockStorageBlobGetter := nsgpeektest.NewFakeStorageBlobGetter(fakeBlob, []string{
		fmt.Sprintf("/subscriptions/xxxx/resourceGroups/xxxx/providers/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=05/d=01/h=12/m=00/macAddress=abc/PT1H.json", fakeNsgName),
		fmt.Sprintf("/SUBSCRIPTIONS/xxxx/RESOURCEGROUPS/xxxx/PROVIDERS/microsoft.network/NETWORKSECURITYGROUPS/%v/y=2022/m=05/d=01/h=12/m=00/macAddress=abc/PT1H.json", "blah"),
		"abc",
		"asdji2wd29jasdjio2/kla0/!?!(*",
		"123/\\///",
	})

	finder := Finder{
		storageBlobGetter: mockStorageBlobGetter,
		nsgName:           fakeNsgName,
	}

	overrideGetBlobUrl(fakeBlobUrl)

	t.Run("TestFindLatestGetsNewestBlobWithCorrectPrefix", func(t *testing.T) {
		expectedPrefix := fmt.Sprintf("/subscriptions/xxxx/resourceGroups/xxxx/providers/microsoft.network/NETWORKSECURITYGROUPS/%v/", fakeNsgName)
		go finder.FindLatest(blobCh, errCh, time.Second*3)
		waitForBlob(t, blobCh, errCh, fakeBlob, time.Second*5)

		if mockStorageBlobGetter.NewestBlobSearchPrefix != expectedPrefix {
			t.Errorf("incorrect blob prefix searched.  expected: '%v', got '%v'", expectedPrefix, mockStorageBlobGetter.NewestBlobSearchPrefix)
		}
	})

	t.Run("TestFindLatestSendsNewBlob", func(t *testing.T) {
		go finder.FindLatest(blobCh, errCh, time.Second*2)
		waitForBlob(t, blobCh, errCh, fakeBlob, time.Second*5)

		// change the newest blob
		fakeNewBlob := new(azure.Blob)
		mockStorageBlobGetter.NewestBlob = fakeNewBlob
		overrideGetBlobUrl(fakeBlobUrl + "/new")
		waitForBlob(t, blobCh, errCh, fakeNewBlob, time.Second*5)
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
