// SPDX-License-Identifier: GPL-3.0-or-later

package broadcom_hba

import (
	"fmt"
	"strings"

	"github.com/netdata/netdata/go/go.d.plugin/agent/module"
)

const (
	prioDeviceROCTemperature = module.Priority + iota
)

var deviceChartsTmpl = module.Charts{
	deviceROCTemperatureChartTmpl.Copy(),
}

var deviceROCTemperatureChartTmpl = module.Chart{
	ID:       "device_%s_temperature",
	Title:    "ROC temperature",
	Units:    "celsius",
	Fam:      "temperature",
	Ctx:      "hba.device_temperature",
	Priority: prioDeviceROCTemperature,
	Dims: module.Dims{
		{ID: "device_%s_temperature", Name: "temperature"},
	},
}

func (bh *BroadcomHBA) addDeviceCharts(device string) {
	charts := deviceChartsTmpl.Copy()

	for _, chart := range *charts {
		chart.ID = fmt.Sprintf(chart.ID, device)
		chart.Labels = []module.Label{
			{Key: "device", Value: device},
		}
		for _, dim := range chart.Dims {
			dim.ID = fmt.Sprintf(dim.ID, device)
		}
	}

	if err := bh.Charts().Add(*charts...); err != nil {
		bh.Warning(err)
	}
}

func (bh *BroadcomHBA) removeDeviceCharts(device string) {
	px := fmt.Sprintf("device_%s", device)

	for _, chart := range *bh.Charts() {
		if strings.HasPrefix(chart.ID, px) {
			chart.MarkRemove()
			chart.MarkNotCreated()
		}
	}
}
