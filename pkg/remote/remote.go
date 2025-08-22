package remote

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"

	"github.com/k8s-school/kadmiral/resources"
)

// It can be better to merge errCh and outCh into a single channel of a struct type,
// so each goroutine sends both output and error together. This avoids race conditions
// and ensures outputs/errors are paired per host.

type RunResult struct {
	Host   string
	Output string
	Err    error
}

func RunParallel(hosts []string, user, key, script string, deps []string) ([]string, []error) {
	slog.Info("run script", "script", script, "hosts", hosts)
	var wg sync.WaitGroup
	resultCh := make(chan RunResult, len(hosts))
	for _, h := range hosts {
		host := h
		wg.Add(1)
		go func() {
			defer wg.Done()

			for _, file := range append([]string{script}, deps...) {
				if err := uploadScript(host, user, key, file); err != nil {
					slog.Error("failed to upload script", "host", host, "script", file, "err", err)
					resultCh <- RunResult{Host: host, Output: "", Err: err}
					return
				}
			}

			command := []string{"bash", "/tmp/kubeadm/" + script}
			out, err := RunCommand(host, user, key, command, deps)
			msg := strings.TrimSpace(string(out))
			resultCh <- RunResult{
				Host:   host,
				Output: msg,
				Err:    err,
			}
		}()
	}
	wg.Wait()
	close(resultCh)

	var outputs []string
	var errs []error
	for res := range resultCh {
		outputs = append(outputs, res.Output)
		if res.Err != nil {
			errs = append(errs, fmt.Errorf("ssh %s: %v", res.Host, res.Err))
		} else {
			errs = append(errs, nil)
		}

	}
	return outputs, errs
}

func RunCommand(host, user, key string, command []string, deps []string) ([]byte, error) {
	sshArgs := []string{}
	if key != "" {
		sshArgs = append(sshArgs, "-i", key)
	}
	sshArgs = append(sshArgs, fmt.Sprintf("%s@%s", user, host))
	sshArgs = append(sshArgs, command...)
	slog.Info("exec script", "host", host, "command", command)
	cmd := exec.Command("ssh", sshArgs...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		msg := strings.TrimSpace(string(out))
		slog.Error("script failed", "host", host, "err", err, "output", msg)
	} else {
		slog.Info("script complete", "host", host)
	}
	return out, err
}

func uploadScript(host string, user, key, script string) error {

	slog.Debug("copy script", "host", host, "path", script)
	remotePath := "/tmp/kubeadm/" + script
	copyCmd := exec.Command(
		"ssh",
		append(
			func() []string {
				if key != "" {
					return []string{"-i", key}
				}
				return []string{}
			}(),
			fmt.Sprintf("%s@%s", user, host),
			fmt.Sprintf("mkdir -p /tmp/kubeadm/ubuntu && cat > %s", remotePath),
		)...,
	)
	slog.Debug("copy command", "cmd", copyCmd.String())
	data, err := resources.Fs.ReadFile(script)
	if err != nil {
		slog.Error("failed to read script", "script", script, "err", err)
		return fmt.Errorf("read script %s: %v", script, err)
	}

	copyCmd.Stdin = strings.NewReader(string(data))
	if out, err := copyCmd.CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(out))
		slog.Error("failed to copy script", "host", host, "err", err, "output", msg)
		return fmt.Errorf("copy script to %s: %v: %s", host, err, msg)
	}
	slog.Debug("script copied", "host", host, "path", remotePath)
	return nil
}
