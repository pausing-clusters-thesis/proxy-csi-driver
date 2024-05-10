package main

import (
	"flag"
	"os"

	proxycsidriver "github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/cmd/proxy-csi-driver"
	"github.com/scylladb/scylla-operator/pkg/genericclioptions"
	"k8s.io/component-base/cli"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(flag.CommandLine)

	command := proxycsidriver.NewCommand(genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})

	code := cli.Run(command)
	os.Exit(code)
}
