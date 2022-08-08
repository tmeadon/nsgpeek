package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/briandowns/spinner"
	"github.com/tmeadon/nsgpeek/pkg/blobreader"
	"github.com/tmeadon/nsgpeek/pkg/flowwriter"
	"github.com/tmeadon/nsgpeek/pkg/logblobfinder"
)

var (
	subscriptionID           string = ""
	nsgRg                    string = "nsg-view"
	nsgName                  string = "nsg-view"
	flowLogBlobContainerName string = "insights-logs-networksecuritygroupflowevent"
)

func main() {
	// subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set")
	}

	// auth
	log.Print("authenticating")

	cred, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	finder, err := logblobfinder.NewLogBlobFinder(subscriptionID, nsgName, log.Default(), context.Background(), cred)
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

	blobReader := blobreader.NewBlobReader(blob, dataCh, errCh)
	go blobReader.Stream(streamStopCh)

	flowWriter := flowwriter.NewFlowWriter(os.Stdout)
	spinner := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
	spinner.Prefix = "waiting for nsg logs...  "

	for {
		spinner.Start()
		select {
		case blob := <-blobCh:
			log.Printf("new blob found: %v", blob.URL())
			streamStopCh <- true
			blobReader = blobreader.NewBlobReader(blob, dataCh, errCh)
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
