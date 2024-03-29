//go:build dev
// +build dev

package brocoindnotify

import (
	"fmt"
	"time"

	"github.com/brsuite/brond/chaincfg/chainhash"
	"github.com/btcsuite/btcwallet/chain"
	"github.com/brolightningnetwork/broln/chainntnfs"
)

// UnsafeStart starts the notifier with a specified best height and optional
// best hash. Its bestBlock and txNotifier are initialized with bestHeight and
// optionally bestHash. The parameter generateBlocks is necessary for the
// brocoind notifier to ensure we drain all notifications up to syncHeight,
// since if they are generated ahead of UnsafeStart the chainConn may start up
// with an outdated best block and miss sending ntfns. Used for testing.
func (b *BrocoindNotifier) UnsafeStart(bestHeight int32, bestHash *chainhash.Hash,
	syncHeight int32, generateBlocks func() error) error {

	// Connect to brocoind, and register for notifications on connected,
	// and disconnected blocks.
	if err := b.chainConn.Start(); err != nil {
		return err
	}
	if err := b.chainConn.NotifyBlocks(); err != nil {
		return err
	}

	b.txNotifier = chainntnfs.NewTxNotifier(
		uint32(bestHeight), chainntnfs.ReorgSafetyLimit,
		b.confirmHintCache, b.spendHintCache,
	)

	if generateBlocks != nil {
		// Ensure no block notifications are pending when we start the
		// notification dispatcher goroutine.

		// First generate the blocks, then drain the notifications
		// for the generated blocks.
		if err := generateBlocks(); err != nil {
			return err
		}

		timeout := time.After(60 * time.Second)
	loop:
		for {
			select {
			case ntfn := <-b.chainConn.Notifications():
				switch update := ntfn.(type) {
				case chain.BlockConnected:
					if update.Height >= syncHeight {
						break loop
					}
				}
			case <-timeout:
				return fmt.Errorf("unable to catch up to height %d",
					syncHeight)
			}
		}
	}

	// Run notificationDispatcher after setting the notifier's best block
	// to avoid a race condition.
	b.bestBlock = chainntnfs.BlockEpoch{Height: bestHeight, Hash: bestHash}
	if bestHash == nil {
		hash, err := b.chainConn.GetBlockHash(int64(bestHeight))
		if err != nil {
			return err
		}
		b.bestBlock.Hash = hash
	}

	b.wg.Add(1)
	go b.notificationDispatcher()

	return nil
}
