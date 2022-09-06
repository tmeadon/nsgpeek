package cli

import (
	"context"
	"log"
	"time"

	"github.com/briandowns/spinner"
	"github.com/tmeadon/nsgpeek/pkg/azure"
	"github.com/tmeadon/nsgpeek/pkg/blobreader"
	"github.com/tmeadon/nsgpeek/pkg/logblobfinder"
)

type StreamCmd struct {
	commonArgs
}

func (s *StreamCmd) Run(ctx *cliContext) error {
	finder, err := logblobfinder.NewLogBlobFinder(subs, s.NsgName, context.Background(), cred)
	if err != nil {
		return err
	}

	blobCh := make(chan (*azure.Blob))
	dataCh := make(chan ([][]byte))
	streamStopCh := make(chan (bool))
	errCh := make(chan (error))

	var blob *azure.Blob
	go finder.FindLatest(blobCh, errCh, time.Second*10)

	select {
	case err := <-errCh:
		return err
	case blob = <-blobCh:
	}

	blobReader := blobreader.NewBlobReader(blob, dataCh, errCh)
	go blobReader.Stream(streamStopCh)

	writers, err := initWriterGroup(s.commonArgs)
	if err != nil {
		return err
	}

	spinner := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
	spinner.Prefix = "waiting for nsg logs...  "

	for {
		spinner.Start()

		select {
		case blob := <-blobCh:
			streamStopCh <- true
			blobReader = blobreader.NewBlobReader(blob, dataCh, errCh)
			go blobReader.Stream(streamStopCh)

		case data := <-dataCh:
			spinner.Stop()
			for _, d := range data {
				writers.WriteFlowBlock(d)
			}
			writers.Flush()
			spinner.Start()

		case err := <-errCh:
			log.Fatalf("error encountered: %v", err)
		}
	}
}
