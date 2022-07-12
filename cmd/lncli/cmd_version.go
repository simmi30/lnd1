package main

import (
	"fmt"

	"github.com/brolightningnetwork/broln/build"
	"github.com/brolightningnetwork/broln/lnrpc/lnclipb"
	"github.com/brolightningnetwork/broln/lnrpc/verrpc"
	"github.com/urfave/cli"
)

var versionCommand = cli.Command{
	Name:  "version",
	Usage: "Display lncli and broln version info.",
	Description: `
	Returns version information about both lncli and broln. If lncli is unable
	to connect to broln, the command fails but still prints the lncli version.
	`,
	Action: actionDecorator(version),
}

func version(ctx *cli.Context) error {
	ctxc := getContext()
	conn := getClientConn(ctx, false)
	defer conn.Close()

	versions := &lnclipb.VersionResponse{
		Lncli: &verrpc.Version{
			Commit:        build.Commit,
			CommitHash:    build.CommitHash,
			Version:       build.Version(),
			AppMajor:      uint32(build.AppMajor),
			AppMinor:      uint32(build.AppMinor),
			AppPatch:      uint32(build.AppPatch),
			AppPreRelease: build.AppPreRelease,
			BuildTags:     build.Tags(),
			GoVersion:     build.GoVersion,
		},
	}

	client := verrpc.NewVersionerClient(conn)

	brolnVersion, err := client.GetVersion(ctxc, &verrpc.VersionRequest{})
	if err != nil {
		printRespJSON(versions)
		return fmt.Errorf("unable fetch version from broln: %v", err)
	}
	versions.broln = brolnVersion

	printRespJSON(versions)

	return nil
}
