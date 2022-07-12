package brocoindnotify

import (
	"errors"
	"fmt"

	"github.com/brsuite/brond/chaincfg"
	"github.com/btcsuite/btcwallet/chain"
	"github.com/brolightningnetwork/broln/blockcache"
	"github.com/brolightningnetwork/broln/chainntnfs"
)

// createNewNotifier creates a new instance of the ChainNotifier interface
// implemented by BrocoindNotifier.
func createNewNotifier(args ...interface{}) (chainntnfs.ChainNotifier, error) {
	if len(args) != 5 {
		return nil, fmt.Errorf("incorrect number of arguments to "+
			".New(...), expected 5, instead passed %v", len(args))
	}

	chainConn, ok := args[0].(*chain.BrocoindConn)
	if !ok {
		return nil, errors.New("first argument to brocoindnotify.New " +
			"is incorrect, expected a *chain.BrocoindConn")
	}

	chainParams, ok := args[1].(*chaincfg.Params)
	if !ok {
		return nil, errors.New("second argument to brocoindnotify.New " +
			"is incorrect, expected a *chaincfg.Params")
	}

	spendHintCache, ok := args[2].(chainntnfs.SpendHintCache)
	if !ok {
		return nil, errors.New("third argument to brocoindnotify.New " +
			"is incorrect, expected a chainntnfs.SpendHintCache")
	}

	confirmHintCache, ok := args[3].(chainntnfs.ConfirmHintCache)
	if !ok {
		return nil, errors.New("fourth argument to brocoindnotify.New " +
			"is incorrect, expected a chainntnfs.ConfirmHintCache")
	}

	blockCache, ok := args[4].(*blockcache.BlockCache)
	if !ok {
		return nil, errors.New("fifth argument to brocoindnotify.New " +
			"is incorrect, expected a *blockcache.BlockCache")
	}

	return New(chainConn, chainParams, spendHintCache,
		confirmHintCache, blockCache), nil
}

// init registers a driver for the BrondNotifier concrete implementation of the
// chainntnfs.ChainNotifier interface.
func init() {
	// Register the driver.
	notifier := &chainntnfs.NotifierDriver{
		NotifierType: notifierType,
		New:          createNewNotifier,
	}

	if err := chainntnfs.RegisterNotifier(notifier); err != nil {
		panic(fmt.Sprintf("failed to register notifier driver '%s': %v",
			notifierType, err))
	}
}
