package flowwriter

import "testing"

type fakeWriter struct {
	writtenBlocks [][]byte
	flushCount    int
}

func (fw *fakeWriter) WriteFlowBlock(data []byte) error {
	fw.writtenBlocks = append(fw.writtenBlocks, data)
	return nil
}

func (fw *fakeWriter) Flush() {
	fw.flushCount++
}

var (
	writer1 *fakeWriter
	writer2 *fakeWriter
	wg      WriterGroup
)

func TestWriterGroup(t *testing.T) {
	writer1 = new(fakeWriter)
	writer2 = new(fakeWriter)
	wg = *NewWriterGroup(writer1, writer2)
	data1 := "abc123"
	data2 := "321cba"

	t.Run("WritesDataToAllWriters", func(t *testing.T) {
		wg.WriteFlowBlock([]byte(data1))
		wg.WriteFlowBlock([]byte(data2))

		for _, w := range []fakeWriter{*writer1, *writer2} {
			if len(w.writtenBlocks) != 2 {
				t.Fatalf("expected writer to have 2 blocks written, got %v", len(w.writtenBlocks))
			}

			if string(w.writtenBlocks[0]) != data1 {
				t.Errorf("expected first written block to be %v, got %v", data1, string(w.writtenBlocks[0]))
			}

			if string(w.writtenBlocks[1]) != data2 {
				t.Errorf("expected first written block to be %v, got %v", data2, string(w.writtenBlocks[1]))
			}
		}
	})

	t.Run("FlushFlushesAllWriters", func(t *testing.T) {
		wg.Flush()
		wg.Flush()

		for _, w := range []fakeWriter{*writer1, *writer2} {
			if w.flushCount != 2 {
				t.Errorf("expected flush to have been called twice, but was called %v time", w.flushCount)
			}
		}
	})
}
