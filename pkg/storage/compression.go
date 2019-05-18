// Compression helpers to reduce object size in DynamoDB table.
// See https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/bp-use-s3-too.html

package storage

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
)

func compressObj(obj interface{}) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	if _, err := w.Write(data); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decompressObj(data []byte, obj interface{}) error {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}

	return json.NewDecoder(r).Decode(obj)
}
