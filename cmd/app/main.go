package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/briandowns/spinner"
	"github.com/tmeadon/nsgpeek/pkg/auth"
	"github.com/tmeadon/nsgpeek/pkg/blobreader"
	"github.com/tmeadon/nsgpeek/pkg/flowwriter"
	"github.com/tmeadon/nsgpeek/pkg/logblobfinder"
)

var (
	flowLogBlobContainerName string = "insights-logs-networksecuritygroupflowevent"
)

func main() {

	nsgName := flag.String("n", "", "name of the nsg")
	flag.Parse()

	if *nsgName == "" {
		fmt.Printf("ERROR: missing name parameter\n\n")
		flag.Usage()
		os.Exit(1)
	}

	cred, err := auth.GetCredential()
	if err != nil {
		log.Fatal(err)
	}

	subs, err := auth.GetSubscriptions(cred)
	if err != nil {
		log.Fatal(err)
	}

	finder := logblobfinder.NewLogBlobFinder(subs, *nsgName, log.Default(), context.Background(), cred)
	blobCh := make(chan (*azblob.BlockBlobClient))
	dataCh := make(chan ([][]byte))
	streamStopCh := make(chan (bool))
	errCh := make(chan (error))

	var blob *azblob.BlockBlobClient
	go finder.FindLatest(blobCh, errCh)

	select {
	case err := <-errCh:
		log.Fatalf("error encountered: %v", err)
	case blob = <-blobCh:
	}

	blobReader := blobreader.NewBlobReader(blobreader.NewBlobWrapper(blob), dataCh, errCh)
	go blobReader.Stream(streamStopCh)

	flowWriter := flowwriter.NewFlowWriter(os.Stdout)
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
				flowWriter.WriteFlowBlock(d)
			}
			flowWriter.Flush()
			spinner.Start()

		case err := <-errCh:
			log.Fatalf("error encountered: %v", err)
		}
	}
}
