// SPDX-License-Identifier: GPL-3.0-or-later

package broadcom_hba

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"
)

func (bh *broadcomHBA) collect() (map[string]int64, error) {
	if bh.exec == nil {
		return nil, errors.New("storcli is not initialized (nil)")
	}

	now := time.Now()
	if bh.forceListDevices || now.Sub(bh.listDevicesTime) > bh.listDevicesEvery {
		bh.forceListDevices = false
		bh.listDevicesTime = now
		if err := bh.listDevices(); err != nil {
			return nil, err
		}
	}

	mx := make(map[string]int64)

	for path := range bh.devicePaths {
		if err := bh.collectHBADevice(mx, path); err != nil {
			bh.Error(err)
			bh.forceListDevices = true
			continue
		}
	}

	return mx, nil
}

func (bh *BroadcomHBA) collectBroadcomHBADevice(mx map[string]int64, devicePath string) error {
	stats, err := bh.exec.smartLog(devicePath)
	if err != nil {
		return fmt.Errorf("exec storcli show temperature '%s': %v", devicePath, err)
	}

	device := extractDeviceFromPath(devicePath)

	mx["device_"+device+"_temperature"] = int64(float64(parseValue(stats.Temperature)))

	return nil
}

func (bh *BroadcomHBA) listBroadcomHBADevices() error {
	devices, err := bh.exec.list()
	if err != nil {
		return fmt.Errorf("exec storcli show: %v", err)
	}

	seen := make(map[string]bool)
	for _, v := range devices.Devices {
		device := extractDeviceFromPath(v.DevicePath)
		seen[device] = true

		if !bh.devicePaths[v.DevicePath] {
			bh.devicePaths[v.DevicePath] = true
			bh.addDeviceCharts(device)
		}
	}
	for path := range bh.devicePaths {
		device := extractDeviceFromPath(path)
		if !seen[device] {
			delete(bh.devicePaths, device)
			bh.removeDeviceCharts(device)
		}
	}

	return nil
}

func extractDeviceFromPath(devicePath string) string {
	_, name := filepath.Split(devicePath)
	return name
}

func boolToInt(v bool) int64 {
	if v {
		return 1
	}
	return 0
}

func parseValue(s hbaNumber) int64 {
	v, _ := strconv.ParseFloat(string(s), 64)
	return int64(v)
}
