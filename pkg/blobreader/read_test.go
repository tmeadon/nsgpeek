package blobreader

import (
	"reflect"
	"testing"
	"time"

	"github.com/tmeadon/nsgpeek/internal/nsgpeektest"
)

func TestRead(t *testing.T) {
	var outCh chan [][]byte
	var errCh chan error
	var doneCh chan bool
	var blob *nsgpeektest.FakeBlob
	var erroringBlob *nsgpeektest.FakeErroringBlob
	var testBlobReader *BlobReader

	setup := func() {
		outCh = make(chan ([][]byte))
		errCh = make(chan error)
		doneCh = make(chan bool)
		blob = nsgpeektest.NewFakeBlob()
		erroringBlob = nsgpeektest.NewFakeErroringBlob()
		testBlobReader = NewBlobReader(blob, outCh, errCh)
		_ = erroringBlob
	}

	t.Run("AllBlocksSentOneByOne", func(t *testing.T) {
		setup()
		go testBlobReader.Read(doneCh)
		received := make([][]byte, 0)

	read:
		for {
			select {
			case data := <-outCh:
				if len(data) > 1 {
					t.Errorf("expected to receive one block at a time, got %v blocks", len(data))
				}
				received = append(received, data...)
			case err := <-errCh:
				t.Errorf("unexpected error received: %v", err)
			case <-time.After(time.Second * 5):
				t.Errorf("timed out waiting for blocks to be read")
			case <-doneCh:
				break read
			}
		}

		if len(received) != (len(blob.Blocks) - 2) {
			t.Errorf("unexpected number of blocks received. wanted: %v, got %v", len(blob.Blocks), len(received))
		}
	})

	t.Run("AllBlocksGetRead", func(t *testing.T) {
		setup()
		go testBlobReader.Read(doneCh)

	read:
		for {
			select {
			case <-outCh:
			case err := <-errCh:
				t.Errorf("unexpected error received: %v", err)
			case <-time.After(time.Second * 5):
				t.Errorf("timed out waiting for blocks to be read")
			case <-doneCh:
				break read
			}
		}

		if !reflect.DeepEqual(blob.Blocks, blob.BlocksRead) {
			t.Errorf("unexpected set of read blocks, expected %v, got %v", blob.Blocks, blob.BlocksRead)
		}
	})
}
