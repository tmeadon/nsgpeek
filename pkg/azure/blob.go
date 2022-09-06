package azure

import (
	"bytes"
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type Blob struct {
	azblob.BlockBlobClient
	Path string
}

func (b *Blob) GetBlocks() ([]BlobBlock, error) {
	blocks, err := b.GetBlockList(context.Background(), azblob.BlockListTypeAll, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get block list for blob %v: %w", b.Path, err)
	}

	bb := make([]BlobBlock, 0)

	for _, b := range blocks.CommittedBlocks {
		bb = append(bb, BlobBlock{Name: *b.Name, Size: *b.Size})
	}

	return bb, nil
}

func (b *Blob) ReadBlock(block *BlobBlock, blockIndex int64) ([]byte, error) {
	downloadOpts := azblob.BlobDownloadOptions{Count: &block.Size, Offset: &blockIndex}

	blockGet, err := b.Download(context.Background(), &downloadOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to download block %v in blob %v: %w", block.Name, b.Path, err)
	}

	data := &bytes.Buffer{}
	reader := blockGet.Body(&azblob.RetryReaderOptions{})

	_, err = data.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read block %v in blob %v: %w", block.Name, b.Path, err)
	}

	err = reader.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close reader for block %v in blob %v: %w", block.Name, b.Path, err)
	}

	return data.Bytes(), nil
}

type BlobBlock struct {
	Name string
	Size int64
}
