package blobreader

import (
	"fmt"
)

func (br *BlobReader) Read(doneCh chan bool) {
	blocks, err := br.getBlockList()
	if err != nil {
		br.errCh <- fmt.Errorf("failed to get block list: %w", err)
		return
	}

	index := int64(0)

	// iterate through the blocks, skipping the first and the last
	for i := 0; i < (len(blocks.CommittedBlocks) - 1); i++ {
		d, err := br.readBlock(&blocks.CommittedBlocks[i], index)
		if err != nil {
			br.errCh <- fmt.Errorf("failed to read block %v: %w", blocks.CommittedBlocks[i].Name, err)
			return
		}

		if i != 0 {
			br.outCh <- [][]byte{d}
		}

		index = index + *blocks.CommittedBlocks[i].Size
	}

	doneCh <- true
}
