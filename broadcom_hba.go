// SPDX-License-Identifier: GPL-3.0-or-later

package broadcom_hba

import (
	_ "embed"
	"math/rand"

	"github.com/netdata/netdata/go/go.d.plugin/agent/module"
)

//go:embed "config_schema.json"
var configSchema string

func init() {
	module.Register("broadcom_hba", module.Creator{
		JobConfigSchema: configSchema,
		Defaults: module.Defaults{
			UpdateEvery:        module.UpdateEvery,
			AutoDetectionRetry: module.AutoDetectionRetry,
			Priority:           module.Priority,
			Disabled:           true,
		},
		Create: func() module.Module { return New() },
	})
}

func New() *BroadcomHBA {
	return &BroadcomHBA{
		Config: Config{
			BinaryPath: "storcli",
			LegacyBinaryPath: "storcli-legacy",
			Timeout:    web.Duration(time.Second * 2),
		},

		charts:           &module.Charts{},
	}
}

type Config struct {
	UpdateEvery int          `yaml:"update_every" json:"update_every"`
	Timeout     web.Duration `yaml:"timeout" json:"timeout"`
}

type (
	BroadcomHBA struct {
		module.Base
		Config `yaml:",inline" json:""`

		charts *module.Charts

		exec broadcomHBA

		devicePaths      map[string]bool
		listDevicesTime  time.Time
		listDevicesEvery time.Duration
		forceListDevices bool
	}
	broadcomHBA interface {
		list() (*broadcomHBADeviceList, error)
	}
)

func (bh *BroadcomHBA) Configuration() any {
	return bh.Config
}

func (bh *BroadcomHBA) Init() error {
	if err := bh.validateConfig(); err != nil {
		bh.Errorf("config validation: %v", err)
		return err
	}

	v, err := bh.initBroadcomHBACLIExec()
	if err != nil {
		bh.Errorf("init storcli exec: %v", err)
		return err
	}
	bh.exec = v

	return nil
}

func (bh *BroadcomHBA) Check() error {
	mx, err := bh.collect()
	if err != nil {
		n.Error(err)
		return err
	}
	if len(mx) == 0 {
		return errors.New("no metrics collected")
	}
	return nil
}

func (bh *BroadcomHBA) Charts() *module.Charts {
	return bh.charts
}

func (bh *BroadcomHBA) Collect() map[string]int64 {
	mx, err := bh.collect()
	if err != nil {
		bh.Error(err)
	}

	if len(mx) == 0 {
		return nil
	}
	return mx
}

func (bh *BroadcomHBA) Cleanup() {}
