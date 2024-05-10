// Copyright (c) 2024 ScyllaDB.

package main

import (
	"flag"
	"os"

	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/cmd/proxy-csi-driver"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/component-base/cli"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(flag.CommandLine)

	command := proxy_csi_driver.NewCommand(genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})

	code := cli.Run(command)
	os.Exit(code)
}
