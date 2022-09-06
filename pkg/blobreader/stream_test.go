package blobreader

import (
	"errors"
	"testing"
	"time"

	"github.com/tmeadon/nsgpeek/internal/nsgpeektest"
)

func TestStream(t *testing.T) {
	var outCh chan [][]byte
	var errCh chan error
	var stopCh chan bool
	var blob *nsgpeektest.FakeBlob
	var erroringBlob *nsgpeektest.FakeErroringBlob
	var testBlobReader *BlobReader

	setup := func() {
		outCh = make(chan ([][]byte))
		errCh = make(chan (error))
		stopCh = make(chan (bool))
		blob = nsgpeektest.NewFakeBlob()
		erroringBlob = nsgpeektest.NewFakeErroringBlob()
		testBlobReader = NewBlobReader(blob, outCh, errCh)
	}

	t.Run("TestStreamDoesntSendExistingBlocks", func(t *testing.T) {
		setup()
		go testBlobReader.Stream(stopCh)
		select {
		case data := <-outCh:
			if data != nil {
				t.Errorf("stream sent data present before start")
			}
		case <-time.After(time.Second * 1):
		}
	})

	t.Run("TestStreamSendsNewBlocks", func(t *testing.T) {
		setup()
		go testBlobReader.Stream(stopCh)
		time.Sleep(time.Second)
		blob.AddBlock(2)

		select {
		case data := <-outCh:
			if len(data) > 2 {
				t.Errorf("stream sent too much data. blocks expected: %v; blocks received: %v", 2, len(data))
			}

			for i := 0; i < 2; i++ {
				if string(data[i]) != nsgpeektest.FakeBlobData {
					t.Errorf("stream sent the wrong data.  expected: %v; got %v", nsgpeektest.FakeBlobData, string(data[i]))
				}
			}

		case <-time.After(time.Second * 5):
			t.Error("stream didn't send data when new block was written")
		}
	})

	t.Run("TestStreamStopsCorrectly", func(t *testing.T) {
		setup()
		go testBlobReader.Stream(stopCh)

		select {
		case stopCh <- true:
		default:
		}

		blob.AddBlock(1)

		select {
		case <-outCh:
			t.Error("data received when Stream should have stopped")
		case <-time.After(time.Second * 1):
		}
	})

	t.Run("TestStreamSendsErrorsCorrectly", func(t *testing.T) {
		setup()
		br := NewBlobReader(erroringBlob, outCh, errCh)
		go br.Stream(stopCh)

		select {
		case err := <-errCh:
			if !errors.Is(err, nsgpeektest.ErrGetBlockList) {
				t.Errorf("incorrect error type received.  wanted: %v; got: %v", nsgpeektest.ErrGetBlockList, err)
			}
		case <-time.After(time.Second):
			t.Error(("error not received on error channel"))
		}
	})
}
