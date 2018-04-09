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
	exists, err := CheckClusterExists(name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("unable to import cluster: cluster with name %s already exists", name)
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
		dstGenParent, _ := filepath.Split(opts.dstGeneratedAssetsDir)
		_, srcGenFolder := filepath.Split(opts.srcGeneratedAssetsDir)
		if err := os.MkdirAll(dstGenParent, 0700); err != nil {
			return fmt.Errorf("error creating destination %s: %v", dstGenParent, err)
		}
		cpGenCmd := exec.Command("cp", "-rf", opts.srcGeneratedAssetsDir, dstGenParent)
		if err := cpGenCmd.Run(); err != nil {
			return fmt.Errorf("error copying from %s to %s: %v", opts.srcGeneratedAssetsDir, dstGenParent, err)
		}
		//Rename whatever you imported from to generated
		//using `pax` would've been easier, but I didn't want to risk it not being
		//on whatever distro ends up being the target
		intermediate := filepath.Join(dstGenParent, srcGenFolder)
		mvGenCmd := exec.Command("mv", intermediate, opts.dstGeneratedAssetsDir)
		if err := mvGenCmd.Run(); err != nil {
			return fmt.Errorf("error moving intermediary dir from %s to %s: %v", intermediate, opts.dstGeneratedAssetsDir, err)
		}
		fmt.Fprintf(out, "Successfully copied generated dir from %s to %s.\n", opts.srcGeneratedAssetsDir, opts.dstGeneratedAssetsDir)
	}
	if opts.srcRunsDir != "" {
		dstRunsParent, _ := filepath.Split(opts.dstRunsDir)
		_, srcRunsFolder := filepath.Split(opts.srcRunsDir)
		if err := os.MkdirAll(dstRunsParent, 0700); err != nil {
			return fmt.Errorf("error creating destination %s: %v", dstRunsParent, err)
		}
		cpRunsCmd := exec.Command("cp", "-rf", opts.srcRunsDir, dstRunsParent)
		if err := cpRunsCmd.Run(); err != nil {
			return fmt.Errorf("error copying from %s to %s: %v", opts.srcRunsDir, dstRunsParent, err)
		}
		intermediate := filepath.Join(dstRunsParent, srcRunsFolder)
		mvGenCmd := exec.Command("mv", intermediate, opts.dstRunsDir)
		if err := mvGenCmd.Run(); err != nil {
			return fmt.Errorf("error moving intermediary dir from %s to %s: %v", intermediate, opts.dstRunsDir, err)
		}
		fmt.Fprintf(out, "Successfully copied runs dir from %s to %s.\n", opts.srcRunsDir, opts.dstRunsDir)
	}
	dstPlanParent, _ := filepath.Split(opts.dstPlanFilePath)
	if err := os.MkdirAll(dstPlanParent, 0700); err != nil {
		return fmt.Errorf("error creating destination %s: %v", dstPlanParent, err)
	}
	cpPlanCmd := exec.Command("cp", "-rf", opts.srcPlanFilePath, opts.dstPlanFilePath)
	if err := cpPlanCmd.Run(); err != nil {
		return fmt.Errorf("error copying from %s to %s: %v", opts.srcPlanFilePath, opts.dstPlanFilePath, err)
	}
	fmt.Fprintf(out, "Successfully copied plan file from %s to %s.\n", opts.srcPlanFilePath, opts.dstPlanFilePath)

	return nil
}
