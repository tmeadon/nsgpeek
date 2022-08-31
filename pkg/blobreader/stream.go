package blobreader

import (
	"fmt"
	"time"
)

func (br *BlobReader) Stream(stopCh chan (bool)) {
	readPosition, err := br.skipToEnd()
	if err != nil {
		br.errCh <- fmt.Errorf("failed to read to end of blob: %w", err)
		return
	}

	stop := false

	for {
		select {
		case stop = <-stopCh:
		case <-time.After(time.Second * 5):
			pos, err := br.readNewBlocks(int64(readPosition))
			if err != nil {
				br.errCh <- err
				return
			}
			readPosition = pos
		}

		if stop {
			break
		}
	}
}

func (br *BlobReader) skipToEnd() (int64, error) {
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
