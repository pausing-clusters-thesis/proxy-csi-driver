// Copyright (c) 2024 ScyllaDB.

package proxy_csi_driver

import (
	"github.com/scylladb/scylla-operator/pkg/cmdutil"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func NewCommand(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.ReadFlagsFromEnv("DELAYED_CSI_DRIVER_", cmd)
		},
	}

	cmd.AddCommand(NewControllerServerCommand(streams))
	cmd.AddCommand(NewNodeServerCommand(streams))
	cmd.AddCommand(NewWaitCommand(streams))

	cmdutil.InstallKlog(cmd)

	return cmd
}
