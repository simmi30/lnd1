package neutrino_test

import (
	"testing"

	lnwallettest "github.com/brolightningnetwork/broln/lnwallet/test"
)

// TestLightningWallet tests LightningWallet powered by neutrino against our
// suite of interface tests.
func TestLightningWallet(t *testing.T) {
	lnwallettest.TestLightningWallet(t, "neutrino")
}
