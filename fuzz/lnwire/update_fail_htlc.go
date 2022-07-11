//go:build gofuzz
// +build gofuzz

package lnwirefuzz

import (
	"github.com/brolightningnetwork/broln/lnwire"
)

// Fuzz_update_fail_htlc is used by go-fuzz.
func Fuzz_update_fail_htlc(data []byte) int {
	// Prefix with MsgUpdateFailHTLC.
	data = prefixWithMsgType(data, lnwire.MsgUpdateFailHTLC)

	// Pass the message into our general fuzz harness for wire messages!
	return harness(data)
}
