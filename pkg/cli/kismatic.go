package cli

import (
	"io"

	"github.com/spf13/cobra"
)

type installOpts struct {
}

// NewKismaticCommand creates the kismatic command
func NewKismaticCommand(version string, buildDate string, in io.Reader, out, stderr io.Writer) (*cobra.Command, error) {
	opts := &installOpts{}
	cmd := &cobra.Command{
		Use:   "kismatic",
		Short: "kismatic is the main tool for managing your Kubernetes cluster",
		Long: `kismatic is the main tool for managing your Kubernetes cluster
more documentation is available at https://github.com/apprenda/kismatic`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(NewCmdImport(out))
	cmd.AddCommand(NewCmdVersion(buildDate, out))
	cmd.AddCommand(NewCmdVolume(in, out))
	cmd.AddCommand(NewCmdIP(out))
	cmd.AddCommand(NewCmdDashboard(in, out))
	cmd.AddCommand(NewCmdSSH(out))
	cmd.AddCommand(NewCmdInfo(out))
	cmd.AddCommand(NewCmdUpgrade(in, out))
	cmd.AddCommand(NewCmdDiagnostic(out))
	cmd.AddCommand(NewCmdCertificates(out))
	cmd.AddCommand(NewCmdSeedRegistry(out, stderr))
	cmd.AddCommand(NewCmdServer(out))
	cmd.AddCommand(NewCmdPlan(in, out, opts))
	cmd.AddCommand(NewCmdValidate(out, opts))
	cmd.AddCommand(NewCmdApply(out, opts))
	cmd.AddCommand(NewCmdAddNode(out, opts))
	cmd.AddCommand(NewCmdStep(out, opts))
	cmd.AddCommand(NewCmdProvision(in, out, opts))
	cmd.AddCommand(NewCmdDestroy(in, out, opts))

	return cmd, nil
}
