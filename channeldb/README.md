channeldb
==========

[![Build Status](http://img.shields.io/travis/lightningnetwork/broln.svg)](https://travis-ci.org/lightningnetwork/broln) 
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/brolightningnetwork/broln/blob/master/LICENSE)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/brolightningnetwork/broln/channeldb)

The channeldb implements the persistent storage engine for `broln` and
generically a data storage layer for the required state within the Lightning
Network. The backing storage engine is
[boltdb](https://github.com/coreos/bbolt), an embedded pure-go key-value store
based off of LMDB.

The package implements an object-oriented storage model with queries and
mutations flowing through a particular object instance rather than the database
itself. The storage implemented by the objects includes: open channels, past
commitment revocation states, the channel graph which includes authenticated
node and channel announcements, outgoing payments, and invoices

## Installation and Updating

```shell
⛰  go get -u github.com/brolightningnetwork/broln/channeldb
```
