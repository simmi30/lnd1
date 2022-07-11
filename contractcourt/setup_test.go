package contractcourt

import (
	"testing"

	"github.com/brolightningnetwork/broln/kvdb"
)

func TestMain(m *testing.M) {
	kvdb.RunTests(m)
}
