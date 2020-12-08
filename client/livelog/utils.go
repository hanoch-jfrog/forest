package livelog

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"
	"time"
)

func InSlice(values []string, wantedVal string) bool {
	for _, val := range values {
		if val == wantedVal {
			return true
		}
	}
	return false
}

func SliceToCsv(values []string) string {
	var buf bytes.Buffer
	wr := csv.NewWriter(&buf)
	err := wr.Write(values)
	if err != nil {
		return ""
	}

	wr.Flush()
	return strings.TrimSuffix(buf.String(), "\n")
}

func MillisToDuration(timeInMillis int64) time.Duration {
	return time.Duration(timeInMillis) * time.Millisecond
}

type errReader struct {
	err error
}

func newErrReader(err error) *errReader {
	return &errReader{err: err}
}

func (er *errReader) Read(_ []byte) (n int, err error) {
	return 0, er.err
}

type blockingReader struct {
	readerChan <-chan io.Reader
	errChan    <-chan error
}

func newBlockingReader(readerChan <-chan io.Reader, errChan <-chan error) *blockingReader {
	return &blockingReader{
		readerChan: readerChan,
		errChan:    errChan,
	}
}

func (br *blockingReader) Read(p []byte) (int, error) {
	n := 0
	for {
		select {
		case r := <-br.readerChan:
			currN, currErr := r.Read(p)
			n += currN
			if currErr != nil && currErr != io.EOF {
				return n, currErr
			}
		case err := <-br.errChan:
			return n, err
		}
	}
}
