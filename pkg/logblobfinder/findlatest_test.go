package logblobfinder

import (
	"reflect"
	"testing"
	"time"

	"github.com/tmeadon/nsgpeek/pkg/azure"
)

var (
	finder                Finder
	blobCh                chan (*azure.Blob)
	errCh                 chan (error)
	fakeBlob              *azure.Blob
	fakeNewBlob           *azure.Blob
	mockNsgGetter         *fakeNsgGetter
	mockStorageBlobGetter *fakeStorageBlobGetter
	fakeBlobUrl           = "https://path.to/blob"
	fakeStorageResourceId = azure.ResourceId{}
	fakeSubscriptionIds   = []string{
		"abc",
		"def",
		"ghi",
	}
)

type fakeNsgGetter struct {
	SubscriptionsSearched []string
}

func (f *fakeNsgGetter) GetNsgFlowLogStorageId(subscriptionIds []string) (*azure.ResourceId, error) {
	f.SubscriptionsSearched = make([]string, 0)
	f.SubscriptionsSearched = append(f.SubscriptionsSearched, subscriptionIds...)
	return &fakeStorageResourceId, nil
}

type fakeStorageBlobGetter struct {
	NewestBlob        *azure.Blob
	StorageIdSearched *azure.ResourceId
}

func (f *fakeStorageBlobGetter) GetNewestBlob(stgAccId *azure.ResourceId) (*azure.Blob, error) {
	f.StorageIdSearched = stgAccId
	return f.NewestBlob, nil
}

func mockGetBlobUrl(*azure.Blob) string {
	return fakeBlobUrl
}

func TestMain(m *testing.M) {
	fakeBlob = new(azure.Blob)
	blobCh = make(chan (*azure.Blob), 0)
	errCh = make(chan (error), 0)
	mockNsgGetter = new(fakeNsgGetter)
	mockStorageBlobGetter = new(fakeStorageBlobGetter)
	mockStorageBlobGetter.NewestBlob = fakeBlob

	finder = Finder{
		nsgGetter:          mockNsgGetter,
		storageBlobGetter:  mockStorageBlobGetter,
		allSubscriptionIds: fakeSubscriptionIds,
	}

	overrideGetBlobUrl(fakeBlobUrl)
	m.Run()
}

func overrideGetBlobUrl(url string) {
	getBlobUrl = func(*azure.Blob) string {
		return url
	}
}

func TestFindLatestSearchesAllSubscriptions(t *testing.T) {
	go finder.FindLatest(blobCh, errCh, time.Second)
	time.Sleep(time.Second)

	if !reflect.DeepEqual(mockNsgGetter.SubscriptionsSearched, fakeSubscriptionIds) {
		t.Errorf("nsg was not searched for in the correct subscriptions.  expected: %v; got: %v", fakeSubscriptionIds, mockNsgGetter.SubscriptionsSearched)
	}
}

func TestFindLatestSearchesCorrectStorageAccount(t *testing.T) {
	go finder.FindLatest(blobCh, errCh, time.Second)
	time.Sleep(time.Second)

	if mockStorageBlobGetter.StorageIdSearched != &fakeStorageResourceId {
		t.Errorf("expected storage account resource ID wasn't searched")
	}
}

func TestFindLatestSendsCorrectBlob(t *testing.T) {
	blobCh = make(chan (*azure.Blob), 0)
	errCh = make(chan (error), 0)

	go finder.FindLatest(blobCh, errCh, time.Second*3)
	waitForBlob(t, blobCh, errCh, fakeBlob, time.Second*5)

	// change the newest blob
	fakeNewBlob = new(azure.Blob)
	mockStorageBlobGetter.NewestBlob = fakeNewBlob
	overrideGetBlobUrl(fakeBlobUrl + "/new")
	waitForBlob(t, blobCh, errCh, fakeNewBlob, time.Second*5)
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
