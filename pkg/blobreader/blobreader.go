package blobreader

import (
	"bytes"
	"context"
	"fmt"

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
