//go:build gofuzz
// +build gofuzz

package lnwirefuzz

import (
	"github.com/brolightningnetwork/broln/lnwire"
)

// Fuzz_update_add_htlc is used by go-fuzz.
func Fuzz_update_add_htlc(data []byte) int {
	// Prefix with MsgUpdateAddHTLC.
	data = prefixWithMsgType(data, lnwire.MsgUpdateAddHTLC)

	// Pass the message into our general fuzz harness for wire messages!
	return harness(data)
}
