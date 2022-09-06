package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tmeadon/nsgpeek/pkg/azure"
	"github.com/tmeadon/nsgpeek/pkg/blobreader"
	"github.com/tmeadon/nsgpeek/pkg/flowwriter"
	"github.com/tmeadon/nsgpeek/pkg/logblobfinder"
)

type SearchCmd struct {
	commonArgs
	Start time.Time `required:"" help:"Start time (UTC) for the log search in format '2006-01-02 15:04:05'" format:"2006-01-02 15:04:05"`
	End   time.Time `required:"" help:"End time (UTC) for the log search in format '2006-01-02 15:04:05'"  format:"2006-01-02 15:04:05"`
}

func (s *SearchCmd) Run(ctx *cliContext) error {
	finder, err := logblobfinder.NewLogBlobFinder(subs, s.NsgName, context.Background(), cred)
	if err != nil {
		return err
	}

	blobs, err := finder.FindSpecific(s.Start, s.End)
	if err != nil {
		return err
	}

	dataCh := make(chan [][]byte)
	errCh := make(chan error)
	waitCh := make(chan bool)
	go readBlobs(blobs, dataCh, errCh, waitCh)

	writers, err := initWriterGroup(s.commonArgs)
	if err != nil {
		return err
	}

	writers.AddFilter(flowwriter.NewTimeFilter(s.Start, s.End))

read:
	for {
		select {
		case data := <-dataCh:
			for _, d := range data {
				writers.WriteFlowBlock(d)
			}
		case err := <-errCh:
			writers.Flush()
			return fmt.Errorf("error: %w", err)
		case <-waitCh:
			writers.Flush()
			break read
		}

	}

	return nil
}

func readBlobs(blobs []azure.Blob, dataCh chan [][]byte, errCh chan error, waitCh chan bool) {
	var wg sync.WaitGroup

	for i, b := range blobs {
		wg.Add(1)
		go readBlob(i, b, dataCh, errCh, &wg)
	}

	wg.Wait()
	close(waitCh)
}

func readBlob(i int, b azure.Blob, dataCh chan [][]byte, errCh chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	childErrCh := make(chan error)
	blobReader := blobreader.NewBlobReader(&b, dataCh, childErrCh)

	doneCh := make(chan bool)
	go blobReader.Read(doneCh)

	select {
	case <-doneCh:
	case err := <-childErrCh:
		errCh <- fmt.Errorf("failed to read blob %v: %w", b.Path, err)
	case <-time.After(time.Minute):
		errCh <- fmt.Errorf("timed out reading blob %v", b.URL())
	}
}
