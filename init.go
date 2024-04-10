// SPDX-License-Identifier: GPL-3.0-or-later

package broadcom_hba

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (bh *broadcomHBA) validateConfig() error {
	if bh.BinaryPath == "" {
		return errors.New("'binary_path' can not be empty")
	}

	return nil
}

func (bh *broadcomHBA) initBroadcomHBACLIExec() (BroadcomHBACLIExec, error) {
	if exePath, err := os.Executable(); err == nil {
		ndsudoPath := filepath.Join(filepath.Dir(exePath), "ndsudo")

		if fi, err := os.Stat(ndsudoPath); err == nil {
			// executable by owner or group
			if fi.Mode().Perm()&0110 != 0 {
				n.Debug("using ndsudo")
				return &BroadcomHBACLIExec{
					ndsudoPath: ndsudoPath,
					timeout:    n.Timeout.Duration(),
				}, nil
			}
		}
	}

	// TODO: remove after next minor release of Netdata (latest is v1.44.0)
	// can't remove now because it will break "from source + stable channel" installations
	bhPath, err := exec.LookPath(bh.BinaryPath)
	if err != nil {
		return nil, err
	}

	var sudoPath string
	if os.Getuid() != 0 {
		sudoPath, err = exec.LookPath("sudo")
		if err != nil {
			return nil, err
		}
	}

	if sudoPath != "" {
		ctx1, cancel1 := context.WithTimeout(context.Background(), bh.Timeout.Duration())
		defer cancel1()

		if _, err := exec.CommandContext(ctx1, sudoPath, "-n", "-v").Output(); err != nil {
			return nil, fmt.Errorf("can not run sudo on this host: %v", err)
		}

		ctx2, cancel2 := context.WithTimeout(context.Background(), n.Timeout.Duration())
		defer cancel2()

		if _, err := exec.CommandContext(ctx2, sudoPath, "-n", "-l", nvmePath).Output(); err != nil {
			return nil, fmt.Errorf("can not run '%s' with sudo: %v", n.BinaryPath, err)
		}
	}

	return &BroadcomHBACLIExec{
		sudoPath: sudoPath,
		bhPath:   bhPath,
		timeout:  bh.Timeout.Duration(),
	}, nil
}
