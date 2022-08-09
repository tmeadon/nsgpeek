package blobreader

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type Blob interface {
	Download(ctx context.Context, options *azblob.BlobDownloadOptions) (BlobDownloadResponse, error)
	GetBlockList(ctx context.Context, listType azblob.BlockListType, options *azblob.BlockBlobGetBlockListOptions) (BlobWrapperGetBlockListResponse, error)
}

type BlobReader struct {
	blob  Blob
	outCh chan ([][]byte)
	errCh chan error
}

func NewBlobReader(blob Blob, outCh chan ([][]byte), errCh chan (error)) *BlobReader {
	return &BlobReader{
		blob:  blob,
		outCh: outCh,
		errCh: errCh,
	}
}

func (br *BlobReader) Stream(stop chan (bool)) {
	readPosition, err := br.readToEnd()
	if err != nil {
		br.errCh <- fmt.Errorf("failed to read to end of blob: %w", err)
		return
	}

	for {
		select {
		case _ = <-stop:
			log.Print("stop")
			break
		case <-time.After(time.Second * 5):
			pos, err := br.readNewBlocks(int64(readPosition))
			if err != nil {
				br.errCh <- err
				return
			}
			readPosition = pos
		}
	}
}

func (br *BlobReader) readToEnd() (int64, error) {
	blocks, err := br.getBlockList()
	if err != nil {
		return 0, err
	}

	index := int64(0)

	for i := 0; i < (len(blocks.CommittedBlocks) - 1); i++ {
		index = index + *blocks.CommittedBlocks[i].Size
	}

	return index, nil
}

func (br *BlobReader) readNewBlocks(offset int64) (int64, error) {
	blocks, err := br.getBlockList()
	if err != nil {
		return 0, err
	}

	index := int64(0)
	var data [][]byte

	// iterate through the blocks, skipping the first and the last
	for i := 0; i < (len(blocks.CommittedBlocks) - 1); i++ {
		if index >= offset {
			d, err := br.readBlock(&blocks.CommittedBlocks[i], index)
			if err != nil {
				return 0, err
			}

			if i != 0 {
				data = append(data, d)
			}
		}

		index = index + *blocks.CommittedBlocks[i].Size
	}

	if len(data) > 0 {
		br.outCh <- data
	}

	return index, nil
}

func (br *BlobReader) getBlockList() (*BlobWrapperGetBlockListResponse, error) {
	blocks, err := br.blob.GetBlockList(context.Background(), azblob.BlockListTypeAll, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get block list: %w", err)
	}

	return &blocks, nil
}

func (br *BlobReader) readBlock(b *BlobWrapperBlock, blockIndex int64) ([]byte, error) {
	downloadOpts := azblob.BlobDownloadOptions{Count: b.Size, Offset: &blockIndex}

	blockGet, err := br.blob.Download(context.Background(), &downloadOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to download block: %w", err)
	}

	data := &bytes.Buffer{}
	reader := blockGet.Body(&azblob.RetryReaderOptions{})

	_, err = data.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read block: %w", err)
	}

	err = reader.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close reader: %w", err)
	}

	return data.Bytes(), nil
}
