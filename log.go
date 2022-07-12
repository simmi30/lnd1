package broln

import (
	"github.com/brsuite/brond/connmgr"
	"github.com/btcsuite/btclog"
	"github.com/lightninglabs/neutrino"
	sphinx "github.com/brolightningnetwork/lightning-onion"
	"github.com/brolightningnetwork/broln/autopilot"
	"github.com/brolightningnetwork/broln/build"
	"github.com/brolightningnetwork/broln/chainntnfs"
	"github.com/brolightningnetwork/broln/chainreg"
	"github.com/brolightningnetwork/broln/chanacceptor"
	"github.com/brolightningnetwork/broln/chanbackup"
	"github.com/brolightningnetwork/broln/chanfitness"
	"github.com/brolightningnetwork/broln/channeldb"
	"github.com/brolightningnetwork/broln/channelnotifier"
	"github.com/brolightningnetwork/broln/cluster"
	"github.com/brolightningnetwork/broln/contractcourt"
	"github.com/brolightningnetwork/broln/discovery"
	"github.com/brolightningnetwork/broln/funding"
	"github.com/brolightningnetwork/broln/healthcheck"
	"github.com/brolightningnetwork/broln/htlcswitch"
	"github.com/brolightningnetwork/broln/invoices"
	"github.com/brolightningnetwork/broln/lnrpc/autopilotrpc"
	"github.com/brolightningnetwork/broln/lnrpc/chainrpc"
	"github.com/brolightningnetwork/broln/lnrpc/invoicesrpc"
	"github.com/brolightningnetwork/broln/lnrpc/routerrpc"
	"github.com/brolightningnetwork/broln/lnrpc/signrpc"
	"github.com/brolightningnetwork/broln/lnrpc/verrpc"
	"github.com/brolightningnetwork/broln/lnrpc/walletrpc"
	"github.com/brolightningnetwork/broln/lnwallet"
	"github.com/brolightningnetwork/broln/lnwallet/btcwallet"
	"github.com/brolightningnetwork/broln/lnwallet/chancloser"
	"github.com/brolightningnetwork/broln/lnwallet/chanfunding"
	"github.com/brolightningnetwork/broln/lnwallet/rpcwallet"
	"github.com/brolightningnetwork/broln/monitoring"
	"github.com/brolightningnetwork/broln/netann"
	"github.com/brolightningnetwork/broln/peer"
	"github.com/brolightningnetwork/broln/peernotifier"
	"github.com/brolightningnetwork/broln/routing"
	"github.com/brolightningnetwork/broln/routing/localchans"
	"github.com/brolightningnetwork/broln/rpcperms"
	"github.com/brolightningnetwork/broln/signal"
	"github.com/brolightningnetwork/broln/sweep"
	"github.com/brolightningnetwork/broln/tor"
	"github.com/brolightningnetwork/broln/watchtower"
	"github.com/brolightningnetwork/broln/watchtower/wtclient"
)

// replaceableLogger is a thin wrapper around a logger that is used so the
// logger can be replaced easily without some black pointer magic.
type replaceableLogger struct {
	btclog.Logger
	subsystem string
}

// Loggers can not be used before the log rotator has been initialized with a
// log file. This must be performed early during application startup by
// calling InitLogRotator() on the main log writer instance in the config.
var (
	// brolnPkgLoggers is a list of all broln package level loggers that are
	// registered. They are tracked here so they can be replaced once the
	// SetupLoggers function is called with the final root logger.
	brolnPkgLoggers []*replaceableLogger

	// addbrolnPkgLogger is a helper function that creates a new replaceable
	// main broln package level logger and adds it to the list of loggers that
	// are replaced again later, once the final root logger is ready.
	addbrolnPkgLogger = func(subsystem string) *replaceableLogger {
		l := &replaceableLogger{
			Logger:    build.NewSubLogger(subsystem, nil),
			subsystem: subsystem,
		}
		brolnPkgLoggers = append(brolnPkgLoggers, l)
		return l
	}

	// Loggers that need to be accessible from the broln package can be placed
	// here. Loggers that are only used in sub modules can be added directly
	// by using the addSubLogger method. We declare all loggers so we never
	// run into a nil reference if they are used early. But the SetupLoggers
	// function should always be called as soon as possible to finish
	// setting them up properly with a root logger.
	ltndLog = addbrolnPkgLogger("LTND")
	rpcsLog = addbrolnPkgLogger("RPCS")
	srvrLog = addbrolnPkgLogger("SRVR")
	atplLog = addbrolnPkgLogger("ATPL")
)

// genSubLogger creates a logger for a subsystem. We provide an instance of
// a signal.Interceptor to be able to shutdown in the case of a critical error.
func genSubLogger(root *build.RotatingLogWriter,
	interceptor signal.Interceptor) func(string) btclog.Logger {

	// Create a shutdown function which will request shutdown from our
	// interceptor if it is listening.
	shutdown := func() {
		if !interceptor.Listening() {
			return
		}

		interceptor.RequestShutdown()
	}

	// Return a function which will create a sublogger from our root
	// logger without shutdown fn.
	return func(tag string) btclog.Logger {
		return root.GenSubLogger(tag, shutdown)
	}
}

// SetupLoggers initializes all package-global logger variables.
func SetupLoggers(root *build.RotatingLogWriter, interceptor signal.Interceptor) {
	genLogger := genSubLogger(root, interceptor)

	// Now that we have the proper root logger, we can replace the
	// placeholder broln package loggers.
	for _, l := range brolnPkgLoggers {
		l.Logger = build.NewSubLogger(l.subsystem, genLogger)
		SetSubLogger(root, l.subsystem, l.Logger)
	}

	// Some of the loggers declared in the main broln package are also used
	// in sub packages.
	signal.UseLogger(ltndLog)
	autopilot.UseLogger(atplLog)

	AddSubLogger(root, "LNWL", interceptor, lnwallet.UseLogger)
	AddSubLogger(root, "DISC", interceptor, discovery.UseLogger)
	AddSubLogger(root, "NTFN", interceptor, chainntnfs.UseLogger)
	AddSubLogger(root, "CHDB", interceptor, channeldb.UseLogger)
	AddSubLogger(root, "HSWC", interceptor, htlcswitch.UseLogger)
	AddSubLogger(root, "CMGR", interceptor, connmgr.UseLogger)
	AddSubLogger(root, "BTCN", interceptor, neutrino.UseLogger)
	AddSubLogger(root, "CNCT", interceptor, contractcourt.UseLogger)
	AddSubLogger(root, "UTXN", interceptor, contractcourt.UseNurseryLogger)
	AddSubLogger(root, "BRAR", interceptor, contractcourt.UseBreachLogger)
	AddSubLogger(root, "SPHX", interceptor, sphinx.UseLogger)
	AddSubLogger(root, "SWPR", interceptor, sweep.UseLogger)
	AddSubLogger(root, "SGNR", interceptor, signrpc.UseLogger)
	AddSubLogger(root, "WLKT", interceptor, walletrpc.UseLogger)
	AddSubLogger(root, "ARPC", interceptor, autopilotrpc.UseLogger)
	AddSubLogger(root, "INVC", interceptor, invoices.UseLogger)
	AddSubLogger(root, "NANN", interceptor, netann.UseLogger)
	AddSubLogger(root, "WTWR", interceptor, watchtower.UseLogger)
	AddSubLogger(root, "NTFR", interceptor, chainrpc.UseLogger)
	AddSubLogger(root, "IRPC", interceptor, invoicesrpc.UseLogger)
	AddSubLogger(root, "CHNF", interceptor, channelnotifier.UseLogger)
	AddSubLogger(root, "CHBU", interceptor, chanbackup.UseLogger)
	AddSubLogger(root, "PROM", interceptor, monitoring.UseLogger)
	AddSubLogger(root, "WTCL", interceptor, wtclient.UseLogger)
	AddSubLogger(root, "PRNF", interceptor, peernotifier.UseLogger)
	AddSubLogger(root, "CHFD", interceptor, chanfunding.UseLogger)
	AddSubLogger(root, "PEER", interceptor, peer.UseLogger)
	AddSubLogger(root, "CHCL", interceptor, chancloser.UseLogger)

	AddSubLogger(root, routing.Subsystem, interceptor, routing.UseLogger, localchans.UseLogger)
	AddSubLogger(root, routerrpc.Subsystem, interceptor, routerrpc.UseLogger)
	AddSubLogger(root, chanfitness.Subsystem, interceptor, chanfitness.UseLogger)
	AddSubLogger(root, verrpc.Subsystem, interceptor, verrpc.UseLogger)
	AddSubLogger(root, healthcheck.Subsystem, interceptor, healthcheck.UseLogger)
	AddSubLogger(root, chainreg.Subsystem, interceptor, chainreg.UseLogger)
	AddSubLogger(root, chanacceptor.Subsystem, interceptor, chanacceptor.UseLogger)
	AddSubLogger(root, funding.Subsystem, interceptor, funding.UseLogger)
	AddSubLogger(root, cluster.Subsystem, interceptor, cluster.UseLogger)
	AddSubLogger(root, rpcperms.Subsystem, interceptor, rpcperms.UseLogger)
	AddSubLogger(root, tor.Subsystem, interceptor, tor.UseLogger)
	AddSubLogger(root, btcwallet.Subsystem, interceptor, btcwallet.UseLogger)
	AddSubLogger(root, rpcwallet.Subsystem, interceptor, rpcwallet.UseLogger)
}

// AddSubLogger is a helper method to conveniently create and register the
// logger of one or more sub systems.
func AddSubLogger(root *build.RotatingLogWriter, subsystem string,
	interceptor signal.Interceptor, useLoggers ...func(btclog.Logger)) {

	// genSubLogger will return a callback for creating a logger instance,
	// which we will give to the root logger.
	genLogger := genSubLogger(root, interceptor)

	// Create and register just a single logger to prevent them from
	// overwriting each other internally.
	logger := build.NewSubLogger(subsystem, genLogger)
	SetSubLogger(root, subsystem, logger, useLoggers...)
}

// SetSubLogger is a helper method to conveniently register the logger of a sub
// system.
func SetSubLogger(root *build.RotatingLogWriter, subsystem string,
	logger btclog.Logger, useLoggers ...func(btclog.Logger)) {

	root.RegisterSubLogger(subsystem, logger)
	for _, useLogger := range useLoggers {
		useLogger(logger)
	}
}

// logClosure is used to provide a closure over expensive logging operations so
// don't have to be performed when the logging level doesn't warrant it.
type logClosure func() string

// String invokes the underlying function and returns the result.
func (c logClosure) String() string {
	return c()
}

// newLogClosure returns a new closure over a function that returns a string
// which itself provides a Stringer interface so that it can be used with the
// logging system.
func newLogClosure(c func() string) logClosure {
	return logClosure(c)
}
