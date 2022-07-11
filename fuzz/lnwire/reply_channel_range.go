//go:build gofuzz
// +build gofuzz

package lnwirefuzz

import (
	"github.com/brolightningnetwork/broln/lnwire"
)

// Fuzz_reply_channel_range is used by go-fuzz.
func Fuzz_reply_channel_range(data []byte) int {
	// Prefix with MsgReplyChannelRange.
	data = prefixWithMsgType(data, lnwire.MsgReplyChannelRange)

	// Pass the message into our general fuzz harness for wire messages!
	return harness(data)
}
