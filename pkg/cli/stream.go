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
	log.Print("creating blob finder")
	finder, err := logblobfinder.NewLogBlobFinder(subs, s.NsgName, context.Background(), cred)
	if err != nil {
		return err
	}

	log.Print("preparing chans")

	blobCh := make(chan (*azure.Blob))
	dataCh := make(chan ([][]byte))
	streamStopCh := make(chan (bool))
	errCh := make(chan (error))

	log.Print("starting spinner")

	// spin := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
	// spin.Prefix = "finding latest nsg flow blob...  "
	// spin.Start()

	log.Print("finding latest")

	var blob *azure.Blob
	go finder.FindLatest(blobCh, errCh, time.Second*10)

	log.Print("stopping spinner")
	// spin.Stop()

	select {
	case err := <-errCh:
		return err
	case blob = <-blobCh:
	}

	log.Print("creating blob reader")

	blobReader := blobreader.NewBlobReader(blob, dataCh, errCh)
	go blobReader.Stream(streamStopCh, time.Second*5)

	log.Print("creating writer group")

	writers, err := initWriterGroup(s.commonArgs)
	if err != nil {
		return err
	}

	spin := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
	spin.Prefix = "waiting for nsg logs...  "

	for {
		spin.Start()
		log.Print("starting loop")

		select {
		case blob := <-blobCh:
			streamStopCh <- true
			blobReader = blobreader.NewBlobReader(blob, dataCh, errCh)
			go blobReader.Stream(streamStopCh, time.Second*5)

		case data := <-dataCh:
			spin.Stop()
			for _, d := range data {
				writers.WriteFlowBlock(d)
			}
			writers.Flush()
			spin.Start()

		case err := <-errCh:
			log.Fatalf("error encountered: %v", err)
		}
	}
}
