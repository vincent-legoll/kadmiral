package cmd

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/k8s-school/kadmiral/pkg/remote"
	"github.com/spf13/cobra"
)

var (
	joinAll    bool
	joinMaster string
	joinToken  string
	joinHash   string
)

var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "join nodes to the cluster",
}

var joinNodeCmd = &cobra.Command{
	Use:   "node [name]",
	Short: "join a worker node",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hosts := AppConfig.WorkerNodes
		printJoinCommand := strings.Split("kubeadm token create --print-join-command --ttl 24h", " ")
		out, err := remote.RunCommand(AppConfig.ControlPlaneNodes[0], AppConfig.SSHUser, AppConfig.SSHKey, printJoinCommand, []string{})
		if err != nil {
			slog.Info("joining nodes", "nodes", hosts, "master", joinMaster)
			return fmt.Errorf("failed to get join command: %v", err)
		}
		commandStr := strings.TrimSpace(string(out))
		commandStr = fmt.Sprintf("sudo %s", commandStr) // prepend sudo to the command
		command := strings.Split(commandStr, " ")
		slog.Debug("WARNING DO NOT PRINT IN LOGjoin command", "command", commandStr)
		_, errs := remote.RunParallelCommand(hosts, AppConfig.SSHUser, AppConfig.SSHKey, command)

		var outErrMsg string
		for i, err := range errs {
			if err != nil {
				slog.Error("failed to join node", "error", err)
				outErrMsg = fmt.Sprintf("%s, node %s failed to join: %v", outErrMsg, hosts[i], err)
			}
		}
		slog.Info("nodes joined", "nodes", hosts)
		if outErrMsg == "" {
			return nil
		} else {
			return fmt.Errorf("failed to join one or more nodes: %s", outErrMsg)
		}
	},
}

func init() {
	joinNodeCmd.Flags().BoolVar(&joinAll, "all", false, "join all nodes")
	joinCmd.AddCommand(joinNodeCmd)
	rootCmd.AddCommand(joinCmd)
}
