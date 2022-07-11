//go:build gofuzz
// +build gofuzz

package lnwirefuzz

import (
	"github.com/brolightningnetwork/broln/lnwire"
)

// Fuzz_update_fail_malformed_htlc is used by go-fuzz.
func Fuzz_update_fail_malformed_htlc(data []byte) int {
	// Prefix with MsgUpdateFailMalformedHTLC.
	data = prefixWithMsgType(data, lnwire.MsgUpdateFailMalformedHTLC)

	// Pass the message into our general fuzz harness for wire messages!
	return harness(data)
}
