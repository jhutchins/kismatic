package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/apprenda/kismatic/pkg/install"

	"github.com/apprenda/kismatic/pkg/ssh"
	"github.com/spf13/cobra"
)

type importOpts struct {
	srcGeneratedAssetsDir string
	dstGeneratedAssetsDir string
	srcRunsDir            string
	dstRunsDir            string
	srcKeyFile            string
	dstKeyFile            string
	srcPlanFilePath       string
	dstPlanFilePath       string
}

// NewCmdImport imports a cluster plan, and potentially a generated or runs dir
func NewCmdImport(out io.Writer) *cobra.Command {
	opts := &importOpts{}
	cmd := &cobra.Command{
		Use:   "import PLAN_FILE_PATH",
		Short: "imports a cluster plan file",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.srcPlanFilePath = args[0]
			fp := install.FilePlanner{File: opts.srcPlanFilePath}
			if !fp.PlanExists() {
				return planFileNotFoundErr{filename: opts.srcPlanFilePath}
			}
			plan, err := fp.Read()
			if err != nil {
				return fmt.Errorf("error reading plan: %v", err)
			}
			clusterName := plan.Cluster.Name
			// Pull destinations from the name
			opts.dstPlanFilePath, opts.dstGeneratedAssetsDir, opts.dstRunsDir = generateDirsFromName(clusterName)
			return doImport(out, clusterName, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.srcKeyFile, "ssh-key", "k", "", "path to the ssh key file")
	cmd.Flags().StringVarP(&opts.srcGeneratedAssetsDir, "generated-assets-dir", "g", "", "path to the directory where assets generated during the installation process were stored")
	cmd.Flags().StringVarP(&opts.srcRunsDir, "runs-dir", "r", "", "path to the directory where artifacts created during the installation process were stored")
	return cmd
}

func doImport(out io.Writer, name string, opts *importOpts) error {

	parent, _ := filepath.Split(opts.dstGeneratedAssetsDir)
	if err := os.MkdirAll(parent, 0700); err != nil {
		return fmt.Errorf("error creating destination %s: %v", parent, err)
	}
	exists, cerr := CheckClusterExists(name)
	if exists {
		fp := install.FilePlanner{File: opts.dstPlanFilePath}
		ofp := install.FilePlanner{File: opts.srcPlanFilePath}
		p, err := fp.Read()
		if err != nil {
			return err
		}
		op, err := ofp.Read()
		if err != nil {
			return err
		}
		if p.Equal(*op) {
			fmt.Fprintf(out, "Identical plan detected in destination %s", opts.dstPlanFilePath)
			return nil
		}
		if cerr != nil {
			return cerr
		}
		return fmt.Errorf("unable to import cluster: cluster already exists")
	}
	if opts.srcKeyFile != "" {
		if err := ssh.ValidUnencryptedPrivateKey(opts.srcKeyFile); err != nil {
			return err
		}
		cpSSHCmd := exec.Command("cp", "-rf", opts.srcKeyFile, opts.dstKeyFile)
		if err := cpSSHCmd.Run(); err != nil {
			return fmt.Errorf("error copying from %s to %s: %v", opts.srcKeyFile, opts.dstKeyFile, err)
		}
		fmt.Fprintf(out, "Successfully copied SSH key from %s to %s.\n", opts.srcKeyFile, opts.dstKeyFile)
	}
	if opts.srcGeneratedAssetsDir != "" {

		_, srcGenFolder := filepath.Split(opts.srcGeneratedAssetsDir)
		cpGenCmd := exec.Command("cp", "-rf", opts.srcGeneratedAssetsDir, parent)
		if err := cpGenCmd.Run(); err != nil {
			return fmt.Errorf("error copying from %s to %s: %v", opts.srcGeneratedAssetsDir, parent, err)
		}
		//Rename whatever you imported from to generated
		//using `pax` would've been easier, but I didn't want to risk it not being
		//on whatever distro ends up being the target
		intermediate := filepath.Join(parent, srcGenFolder)
		mvGenCmd := exec.Command("mv", "-f", intermediate, opts.dstGeneratedAssetsDir)
		if err := mvGenCmd.Run(); err != nil {
			return fmt.Errorf("error moving intermediary dir from %s to %s: %v", intermediate, opts.dstGeneratedAssetsDir, err)
		}
		fmt.Fprintf(out, "Successfully copied generated dir from %s to %s.\n", opts.srcGeneratedAssetsDir, opts.dstGeneratedAssetsDir)
	}
	if opts.srcRunsDir != "" {
		parent, _ := filepath.Split(opts.dstRunsDir)
		_, srcRunsFolder := filepath.Split(opts.srcRunsDir)
		cpRunsCmd := exec.Command("cp", "-rf", opts.srcRunsDir, parent)
		if err := cpRunsCmd.Run(); err != nil {
			return fmt.Errorf("error copying from %s to %s: %v", opts.srcRunsDir, parent, err)
		}
		intermediate := filepath.Join(parent, srcRunsFolder)
		mvGenCmd := exec.Command("mv", "-f", intermediate, opts.dstRunsDir)
		if err := mvGenCmd.Run(); err != nil {
			return fmt.Errorf("error moving intermediary dir from %s to %s: %v", intermediate, opts.dstRunsDir, err)
		}
		fmt.Fprintf(out, "Successfully copied runs dir from %s to %s.\n", opts.srcRunsDir, opts.dstRunsDir)
	}
	cpPlanCmd := exec.Command("cp", "-f", opts.srcPlanFilePath, opts.dstPlanFilePath)
	if err := cpPlanCmd.Run(); err != nil {
		return fmt.Errorf("error copying from %s to %s: %v", opts.srcPlanFilePath, opts.dstPlanFilePath, err)
	}
	fmt.Fprintf(out, "Successfully copied plan file from %s to %s.\n", opts.srcPlanFilePath, opts.dstPlanFilePath)

	return nil
}
