package nsgpeektest

import (
	"errors"

	"github.com/tmeadon/nsgpeek/pkg/azure"
)

var FakeBlobData = "fake"
var ErrGetBlockList error = errors.New("block list get error")

type FakeBlob struct {
	Blocks     []azure.BlobBlock
	BlocksRead []azure.BlobBlock
}

func (f *FakeBlob) ReadBlock(block *azure.BlobBlock, blockIndex int64) ([]byte, error) {
	f.BlocksRead = append(f.BlocksRead, *block)
	return []byte(FakeBlobData), nil
}

func (f *FakeBlob) GetBlocks() ([]azure.BlobBlock, error) {
	return f.Blocks, nil
}

func (f *FakeBlob) AddBlock(count int) {
	for i := 1; i <= count; i++ {
		n := "new"
		s := int64(999)
		f.Blocks = append(f.Blocks, azure.BlobBlock{Name: n, Size: s})
	}
}

func NewFakeBlob() *FakeBlob {
	b := FakeBlob{}
	b.Blocks = makeBlocks()
	return &b
}

type FakeErroringBlob struct {
	blocks []azure.BlobBlock
}

func (f *FakeErroringBlob) ReadBlock(block *azure.BlobBlock, blockIndex int64) ([]byte, error) {
	return []byte(FakeBlobData), nil
}

func (f *FakeErroringBlob) GetBlocks() ([]azure.BlobBlock, error) {
	return f.blocks, ErrGetBlockList
}

func (f *FakeErroringBlob) AddBlock(count int) {
	for i := 1; i <= count; i++ {
		n := "new"
		s := int64(999)
		f.blocks = append(f.blocks, azure.BlobBlock{Name: n, Size: s})
	}
}

func NewFakeErroringBlob() *FakeErroringBlob {
	b := FakeErroringBlob{}
	b.blocks = makeBlocks()
	return &b
}

func makeBlocks() []azure.BlobBlock {
	blocks := make([]azure.BlobBlock, 0)

	blockData := map[string]int64{
		"first":       1,
		"abc":         898,
		"def":         7121,
		"osk":         231,
		"haijcw9ejcw": 12111,
		"last":        1,
	}

	for k, v := range blockData {
		n := k
		s := v
		bl := azure.BlobBlock{
			Name: n,
			Size: s,
		}
		blocks = append(blocks, bl)
	}

	return blocks
}
