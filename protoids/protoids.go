// Package protoids provides protobuf-based ID encoding and decoding routines
package protoids

import (
	"github.com/golang/protobuf/proto"
	"github.com/gotruward/ids"
)

// Decode is a helper method, that takes ID codec, encoded ID and output
// message struct and unmarshals encoded ID into the given message
func Decode(codec ids.IDCodec, encodedID string, pb proto.Message) error {
	raw, err := codec.Decode(encodedID)
	if err != nil {
		return err
	}

	return proto.Unmarshal(raw, pb)
}

// Encode is a helper method, that takes ID codec and message and
// returns string-encoded representation of that message
func Encode(codec ids.IDCodec, pb proto.Message) (string, error) {
	raw, err := proto.Marshal(pb)
	if err != nil {
		return "", err
	}

	return codec.Encode(raw)
}


