package blobreader

import (
	"fmt"
)

func (br *BlobReader) Read(doneCh chan bool) {
	blocks, err := br.blob.GetBlocks()
	if err != nil {
		br.errCh <- fmt.Errorf("failed to get block list: %w", err)
		return
	}

	index := int64(0)

	// iterate through the blocks, skipping the first and the last
	for i := 0; i < len(blocks); i++ {
		d, err := br.blob.ReadBlock(&blocks[i], index)
		if err != nil {
			br.errCh <- err
			return
		}

		if i != 0 && i != (len(blocks)-1) {
			br.outCh <- [][]byte{d}
		}

		index = index + blocks[i].Size
	}

	doneCh <- true
}
