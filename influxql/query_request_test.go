package influxql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodingFormatFromMimeType(t *testing.T) {
	tests := []struct {
		name string
		s    string
		exp  EncodingFormat
	}{
		{s: "application/csv", exp: EncodingFormatAppCSV},
		{s: "text/csv", exp: EncodingFormatTextCSV},
		{s: "application/x-msgpack", exp: EncodingFormatMessagePack},
		{s: "application/json", exp: EncodingFormatJSON},
		{s: "application/other", exp: EncodingFormatUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodingFormatFromMimeType(tt.s)
			assert.Equal(t, tt.exp, got)
		})
	}
}
