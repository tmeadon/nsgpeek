package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/tmeadon/nsgpeek/pkg/azure"
	"github.com/tmeadon/nsgpeek/pkg/blobreader"
	"github.com/tmeadon/nsgpeek/pkg/flowwriter"
	"github.com/tmeadon/nsgpeek/pkg/logblobfinder"
)

type StreamCmd struct {
	NsgName string `required:"" short:"n" help:"Name of the NSG to stream logs from"`
	CsvFile string `help:"Path to a CSV file to stream logs to"`
}

func (s *StreamCmd) Run(ctx *cliContext) error {
	finder := logblobfinder.NewLogBlobFinder(subs, s.NsgName, context.Background(), cred)

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

	blobReader := blobreader.NewBlobReader(blobreader.NewBlobWrapper(blob), dataCh, errCh)
	go blobReader.Stream(streamStopCh)

	flowWriter := flowwriter.NewConsoleWriter(os.Stdout)
	writers := flowwriter.NewWriterGroup(flowWriter)
	addCsvWriter(s.CsvFile, writers)

	spinner := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
	spinner.Prefix = "waiting for nsg logs...  "

	for {
		spinner.Start()

		select {
		case blob := <-blobCh:
			streamStopCh <- true
			blobReader = blobreader.NewBlobReader(blobreader.NewBlobWrapper(blob), dataCh, errCh)
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

func addCsvWriter(path string, wg *flowwriter.WriterGroup) error {
	if path != "" {
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file %v: %w", path, err)
		}

		csvWriter, err := flowwriter.NewCsvFileWriter(file)
		if err != nil {
			return fmt.Errorf("failed to create csv file writer: %w", err)
		}

		wg.AddWriter(csvWriter)
	}
	return nil
}
