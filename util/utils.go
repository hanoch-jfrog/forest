package util

import (
	"bytes"
	"encoding/csv"
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
