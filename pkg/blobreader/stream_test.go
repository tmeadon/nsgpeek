package blobreader

import (
	"errors"
	"testing"
	"time"

	"github.com/tmeadon/nsgpeek/internal/nsgpeektest"
	"github.com/tmeadon/nsgpeek/pkg/azure"
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

	t.Run("DoesntSendExistingBlocks", func(t *testing.T) {
		setup()
		go testBlobReader.Stream(stopCh, time.Second)
		select {
		case data := <-outCh:
			if data != nil {
				t.Errorf("stream sent data present before start")
			}
		case <-time.After(time.Second * 1):
		}
	})

	t.Run("SendsNewBlocks", func(t *testing.T) {
		setup()
		go testBlobReader.Stream(stopCh, time.Second)
		time.Sleep(time.Second)

		newBlocks := []azure.BlobBlock{{Name: "test1", Size: 123}, {Name: "test2", Size: 999}}
		blob.AddBlocks(newBlocks)

		select {
		case data := <-outCh:
			if len(data) > len(newBlocks) {
				t.Errorf("stream sent too much data. blocks expected: %v; blocks received: %v", len(newBlocks), len(data))
			}

			for i := 0; i < len(newBlocks); i++ {
				if string(data[i]) != newBlocks[i].Name {
					t.Errorf("stream sent the wrong data. expected %v, got %v", newBlocks[i].Name, string(data[i]))
				}
			}

		case <-time.After(time.Second * 5):
			t.Error("stream didn't send data when new block was written")
		}
	})

	t.Run("StopsCorrectly", func(t *testing.T) {
		setup()
		go testBlobReader.Stream(stopCh, time.Second)

		select {
		case stopCh <- true:
		default:
		}

		blob.AddBlocks([]azure.BlobBlock{{Name: "test123", Size: 999}})

		select {
		case <-outCh:
			t.Error("data received when Stream should have stopped")
		case <-time.After(time.Second * 1):
		}
	})

	t.Run("SendsErrorsCorrectly", func(t *testing.T) {
		setup()
		br := NewBlobReader(erroringBlob, outCh, errCh)
		go br.Stream(stopCh, time.Second)

		select {
		case err := <-errCh:
			if !errors.Is(err, nsgpeektest.ErrGetBlockList) {
				t.Errorf("incorrect error type received.  wanted: %v; got: %v", nsgpeektest.ErrGetBlockList, err)
			}
		case <-time.After(time.Second):
			t.Error(("error not received on error channel"))
		}
	})

	t.Run("DoesNotSendDuplicateBlocks", func(t *testing.T) {
		setup()
		go testBlobReader.Stream(stopCh, time.Second)
		time.Sleep(time.Second)

		waitForData := func() (received [][]byte) {
			select {
			case received = <-outCh:
			case err := <-errCh:
				t.Fatalf("unexpected error received: %v", err)
			}
			return
		}

		firstNewBlocks := []azure.BlobBlock{{Name: "test1", Size: 123}, {Name: "test2", Size: 999}}
		blob.AddBlocks(firstNewBlocks)

		received := waitForData()
		readBlocks := make(map[string]bool)

		for _, d := range received {
			readBlocks[string(d)] = true
		}

		secondNewBlocks := []azure.BlobBlock{{Name: "test3", Size: 123}, {Name: "test4", Size: 999}}
		blob.AddBlocks(secondNewBlocks)
		time.Sleep(time.Second)

		received = waitForData()

		for _, d := range received {
			if _, ok := readBlocks[string(d)]; ok {
				t.Fatalf("block with data %v has been sent more than once", string(d))
			}
		}
	})
}
