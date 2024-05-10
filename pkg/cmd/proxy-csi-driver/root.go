package proxy_csi_driver

import (
	"github.com/scylladb/scylla-operator/pkg/cmdutil"
	"github.com/scylladb/scylla-operator/pkg/genericclioptions"
	"github.com/spf13/cobra"
)

func NewCommand(streams genericclioptions.IOStreams) *cobra.Command {
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
