package model

import (
	"fmt"
	"strconv"
	"time"

	"github.com/vmihailenco/msgpack"
)

type Timestamp time.Time

func (t Timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(t).Unix()
	stamp := fmt.Sprint(ts)
	return []byte(stamp), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}

	*t = Timestamp(time.Unix(int64(ts), 0))
	return nil
}

func (t Timestamp) EncodeMsgpack(enc *msgpack.Encoder) error {
	ts := time.Time(t).Unix()
	stamp := fmt.Sprint(ts)
	return enc.EncodeString(stamp)
}

func (t *Timestamp) DecodeMsgpack(dec *msgpack.Decoder) error {
	str, err := dec.DecodeString()
	if err != nil {
		// TODO: old cache will fail here :(
		*t = Timestamp{}
		return nil
	}

	ts, err := strconv.Atoi(str)
	if err != nil {
		return err
	}

	*t = Timestamp(time.Unix(int64(ts), 0))
	return nil
}
