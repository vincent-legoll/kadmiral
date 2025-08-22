package cmd

import (
	"fmt"
	"log/slog"

	"github.com/k8s-school/kadmiral/pkg/remote"
	"github.com/k8s-school/kadmiral/resources"
	"github.com/spf13/cobra"
)

var cniCmd = &cobra.Command{
	Use:   "cni",
	Short: "manage CNI plugins",
}

var cniInstallCmd = &cobra.Command{
	Use:   "install [calico|cilium]",
	Short: "install a CNI plugin",
	Example: `  kadmiral cni install cilium
  kadmiral cni install calico (TODO)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		script := fmt.Sprintf("install-%s.sh", name)

		// Check if script exists could be done here
		_, err := resources.Fs.Open(script) // to trigger error if not exists
		if err != nil {
			return fmt.Errorf("CNI plugin %q not supported", name)
		}

		host := AppConfig.ControlPlaneNodes[0]
		slog.Info("installing CNI", "name", name, "node", host)
		// assume CNI is installed on control plane first node
		if _, err := remote.RunParallel([]string{host}, AppConfig.SSHUser, AppConfig.SSHKey, script, nil); err != nil {
			return fmt.Errorf("failed to install CNI: %v", err)
		}
		slog.Info("CNI installed", "name", name)
		return nil
	},
}

func init() {
	cniCmd.AddCommand(cniInstallCmd)
	rootCmd.AddCommand(cniCmd)
}
