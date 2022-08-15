package blobreader

import (
	"context"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/tmeadon/nsgpeek/pkg/azure"
)

type BlobDownloadResponse interface {
	Body(opts *azblob.RetryReaderOptions) io.ReadCloser
}

type BlobWrapperDownloadResponse struct {
	resp BlobDownloadResponse
}

func (r *BlobWrapperDownloadResponse) Body(opts *azblob.RetryReaderOptions) io.ReadCloser {
	return r.resp.Body(opts)
}

type BlobWrapperGetBlockListResponse struct {
	CommittedBlocks []BlobWrapperBlock
}

type BlobWrapperBlock struct {
	Name *string
	Size *int64
}

type BlobWrapper struct {
	blob *azure.Blob
}

func NewBlobWrapper(blob *azure.Blob) *BlobWrapper {
	return &BlobWrapper{blob}
}

func (bw *BlobWrapper) Download(ctx context.Context, options *azblob.BlobDownloadOptions) (BlobDownloadResponse, error) {
	resp, err := bw.blob.Download(ctx, options)
	if err != nil {
		return &BlobWrapperDownloadResponse{}, err
	}

	return &BlobWrapperDownloadResponse{&resp}, nil
}

func (bw *BlobWrapper) GetBlockList(ctx context.Context, listType azblob.BlockListType, options *azblob.BlockBlobGetBlockListOptions) (BlobWrapperGetBlockListResponse, error) {
	resp, err := bw.blob.GetBlockList(ctx, listType, options)
	if err != nil {
		return BlobWrapperGetBlockListResponse{}, nil
	}

	wrappedResp := BlobWrapperGetBlockListResponse{}

	for _, b := range resp.CommittedBlocks {
		wrappedResp.CommittedBlocks = append(wrappedResp.CommittedBlocks, BlobWrapperBlock{b.Name, b.Size})
	}

	return wrappedResp, nil
}
