package blobreader

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var fakeBlobData string = "123abc"
var fakeBlockList []BlobWrapperBlock
var testBlobReader *BlobReader
var outCh chan ([][]byte)
var errCh chan (error)
var stopCh chan (bool)
var errGetBlockList error = errors.New("block list get error")

func addInitialBlocks() {
	blocks := make([]BlobWrapperBlock, 0)

	blockData := map[string]int64{
		"abc":         898,
		"def":         7121,
		"osk":         231,
		"haijcw9ejcw": 12111,
	}

	for k, v := range blockData {
		n := k
		s := v
		b := BlobWrapperBlock{
			Name: &n,
			Size: &s,
		}
		blocks = append(blocks, b)
	}

	fakeBlockList = blocks
}

type fakeBlobDownloadResponse struct{}

func (f *fakeBlobDownloadResponse) Body(opts *azblob.RetryReaderOptions) io.ReadCloser {
	return io.NopCloser(strings.NewReader(fakeBlobData))
}

type fakeBlobWrapper struct{}

func (f *fakeBlobWrapper) Download(ctx context.Context, options *azblob.BlobDownloadOptions) (BlobDownloadResponse, error) {
	return new(fakeBlobDownloadResponse), nil
}

func (f *fakeBlobWrapper) GetBlockList(ctx context.Context, listType azblob.BlockListType, options *azblob.BlockBlobGetBlockListOptions) (BlobWrapperGetBlockListResponse, error) {
	return BlobWrapperGetBlockListResponse{fakeBlockList}, nil
}

type fakeErroringBlobWrapper struct{}

func (f *fakeErroringBlobWrapper) Download(ctx context.Context, options *azblob.BlobDownloadOptions) (BlobDownloadResponse, error) {
	return new(fakeBlobDownloadResponse), nil
}

func (f *fakeErroringBlobWrapper) GetBlockList(ctx context.Context, listType azblob.BlockListType, options *azblob.BlockBlobGetBlockListOptions) (BlobWrapperGetBlockListResponse, error) {
	return BlobWrapperGetBlockListResponse{}, errGetBlockList
}

func TestMain(m *testing.M) {
	addInitialBlocks()
	outCh = make(chan ([][]byte))
	errCh = make(chan (error))
	stopCh = make(chan (bool))
	blob := new(fakeBlobWrapper)
	testBlobReader = NewBlobReader(blob, outCh, errCh)
	m.Run()
}

func TestStreamDoesntSendExistingBlocks(t *testing.T) {
	startStream()
	select {
	case data := <-outCh:
		if data != nil {
			t.Errorf("stream sent data present before start")
		}
	case <-time.After(time.Second * 1):
	}
}

func TestStreamSendsNewBlocks(t *testing.T) {
	startStream()
	time.Sleep(time.Second)
	addBlock(2)

	select {
	case data := <-outCh:
		if len(data) > 2 {
			t.Errorf("stream sent too much data. blocks expected: %v; blocks received: %v", 2, len(data))
		}

		for i := 0; i < 2; i++ {
			if string(data[i]) != fakeBlobData {
				t.Errorf("stream sent the wrong data.  expected: %v; got %v", fakeBlobData, string(data[i]))
			}
		}

	case <-time.After(time.Second * 5):
		t.Error("stream didn't send data when new block was written")
	}
}

func TestStreamStopsCorrectly(t *testing.T) {
	startStream()

	select {
	case stopCh <- true:
	default:
	}

	addBlock(1)

	select {
	case <-outCh:
		t.Error("data received when Stream should have stopped")
	case <-time.After(time.Second * 1):
	}
}

func TestStreamSendsErrorsCorrectly(t *testing.T) {
	erroringBlob := new(fakeErroringBlobWrapper)
	br := NewBlobReader(erroringBlob, outCh, errCh)
	go br.Stream(stopCh)

	select {
	case err := <-errCh:
		if !errors.Is(err, errGetBlockList) {
			t.Errorf("incorrect error type received.  expected: %v; wanted: %v", errGetBlockList, err)
		}
	case <-time.After(time.Second):
		t.Error(("error not received on error channel"))
	}
}

func startStream() {
	go testBlobReader.Stream(stopCh)
}

func addBlock(count int) {
	for i := 1; i <= count; i++ {
		n := "new"
		s := int64(999)
		fakeBlockList = append(fakeBlockList, BlobWrapperBlock{&n, &s})
	}
}
