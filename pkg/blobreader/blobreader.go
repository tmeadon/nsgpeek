package blobreader

import (
	"github.com/tmeadon/nsgpeek/pkg/azure"
)

type Blob interface {
	GetBlocks() ([]azure.BlobBlock, error)
	ReadBlock(block *azure.BlobBlock, blockIndex int64) ([]byte, error)
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
