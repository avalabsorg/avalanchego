package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/node"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/version"
	"github.com/spf13/viper"
)

type nodeProcess struct {
	path     string
	errChan  chan error
	exitCode int
	cmd      *exec.Cmd
}

// Returns a new nodeProcess running the binary at [path].
// Returns an error if the command fails to start.
// When the nodeProcess terminates, the returned error (which may be nil)
// is sent on [n.errChan]
func startNode(path string, args []string, printToStdOut bool) (*nodeProcess, error) {
	fmt.Printf("Starting binary at %s with args %s\n", path, args) // TODO remove
	n := &nodeProcess{
		path:    path,
		cmd:     exec.Command(path, args...), // #nosec G204
		errChan: make(chan error, 1),
	}
	if printToStdOut {
		n.cmd.Stdout = os.Stdout
		n.cmd.Stderr = os.Stderr
	}

	// Start the nodeProcess
	if err := n.cmd.Start(); err != nil {
		return nil, err
	}
	go func() {
		// Wait for the nodeProcess to stop.
		// When it does, set the exit code and send the returned error
		// (which may be nil) to [a.errChain]
		if err := n.cmd.Wait(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				// This code only executes if the exit code is non-zero
				n.exitCode = exitError.ExitCode()
			}
			n.errChan <- err
		}
	}()
	return n, nil
}

func (a *nodeProcess) kill() error {
	if a.cmd.Process == nil {
		return nil
	}
	// Stop printing output from node
	a.cmd.Stdout = ioutil.Discard
	a.cmd.Stderr = ioutil.Discard
	//todo change this to interrupt?
	err := a.cmd.Process.Kill() // todo kill subprocesses
	if err != nil && err != os.ErrProcessDone {
		return fmt.Errorf("failed to kill process: %w", err)
	}
	return nil
}

type binaryManager struct {
	rootPath string
	log      logging.Logger
}

func newBinaryManager(path string, log logging.Logger) *binaryManager {
	return &binaryManager{
		rootPath: path,
		log:      log,
	}
}

// Run two nodes at once: one is a version before the database upgrade and the other after.
// The latter will bootstrap from the former. Its staking port and HTTP port are 2
// greater than the staking/HTTP ports in [v].
// When the new node version is done bootstrapping, both nodes are stopped.
// Returns nil if the new node version successfully bootstrapped.
func (b *binaryManager) runMigration(v *viper.Viper, nodeConfig node.Config) error {
	prevVersionNode, err := b.runPreviousVersion(previousVersion, v)
	if err != nil {
		return fmt.Errorf("couldn't start old version during migration: %w", err)
	}
	defer func() {
		if err := prevVersionNode.kill(); err != nil {
			b.log.Error("error while killing previous version: %w", err)
		}
	}()

	currentVersionNode, err := b.runCurrentVersion(v, true, nodeConfig.NodeID)
	if err != nil {
		return fmt.Errorf("couldn't start current version during migration: %w", err)
	}
	defer func() {
		if err := currentVersionNode.kill(); err != nil {
			b.log.Error("error while killing current version: %w", err)
		}
	}()

	for {
		select {
		case err := <-prevVersionNode.errChan:
			if err != nil {
				return fmt.Errorf("previous version died with exit code %d", prevVersionNode.exitCode)
			}
			if prevVersionNode.exitCode == constants.ExitCodeDoneMigrating {
				return nil
			}
			// TODO restart here
		case err := <-currentVersionNode.errChan:
			if err != nil {
				return fmt.Errorf("current version died with exit code %d", currentVersionNode.exitCode)
			}
		}
	}
}

func (b *binaryManager) runNormal(v *viper.Viper) error {
	node, err := b.runCurrentVersion(v, false, ids.ShortID{})
	if err != nil {
		return fmt.Errorf("couldn't start old version during migration: %w", err)
	}
	return <-node.errChan
}

func (b *binaryManager) runPreviousVersion(prevVersion version.Version, v *viper.Viper) (*nodeProcess, error) {
	binaryPath := getBinaryPath(b.rootPath, prevVersion)
	args := []string{}
	for k, v := range v.AllSettings() {
		if k == "fetch-only" { // TODO replace with const
			continue
		}
		args = append(args, fmt.Sprintf("--%s=%v", k, v))
	}
	return startNode(binaryPath, args, false)
}

func (b *binaryManager) runCurrentVersion(
	v *viper.Viper,
	fetchOnly bool,
	fetchFrom ids.ShortID,
) (*nodeProcess, error) {
	argsMap := v.AllSettings()
	if fetchOnly {
		// TODO use constants for arg names here
		stakingPort, err := strconv.Atoi(argsMap["staking-port"].(string))
		if err != nil {
			return nil, fmt.Errorf("couldn't parse staking port as int: %w", err)
		}
		argsMap["bootstrap-ips"] = fmt.Sprintf("127.0.0.1:%d", stakingPort)
		argsMap["bootstrap-ids"] = fmt.Sprintf("%s%s", constants.NodeIDPrefix, fetchFrom)
		argsMap["staking-port"] = stakingPort + 2

		httpPort, err := strconv.Atoi(argsMap["http-port"].(string))
		if err != nil {
			return nil, fmt.Errorf("couldn't parse staking port as int: %w", err)
		}
		argsMap["http-port"] = httpPort + 2
		argsMap["fetch-only"] = true
	}
	args := []string{}
	for k, v := range argsMap {
		args = append(args, fmt.Sprintf("--%s=%v", k, v))
	}
	binaryPath := getBinaryPath(b.rootPath, currentVersion)
	return startNode(binaryPath, args, true)
}

func getBinaryPath(rootPath string, nodeVersion version.Version) string {
	return fmt.Sprintf(
		"%s/build/avalanchego-%s/avalanchego-inner",
		rootPath,
		nodeVersion,
	)
}
