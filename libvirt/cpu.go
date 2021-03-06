/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package libvirt

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/sandlbn/libvirt-go"
)

var cpuMetricsTypes = []string{"cputime"}

func cpuTimes(ns []string, dom libvirt.VirDomain) (*plugin.PluginMetricType, error) {
	info, err := dom.GetInfo()
	if err != nil {
		return nil, err
	}
	switch {
	case regexp.MustCompile(`^/libvirt/.*/.*/cpu/cputime`).MatchString(joinNamespace(ns)):
		cpuTime := strconv.FormatUint(info.GetCpuTime(), 10)
		if ns[2] == "*" {
			domainName, err := dom.GetName()
			if err != nil {
				return nil, err
			}
			ns[2] = domainName
		}
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      cpuTime,
			Timestamp_: time.Now(),
		}, nil
	case regexp.MustCompile(`^/libvirt/.*/.*/cpu/vcpu/.*/cputime`).MatchString(joinNamespace(ns)):
		nr, err := strconv.Atoi(ns[5])
		if err != nil {
			return nil, err
		}
		metric := getVcpuTime(nr, info, dom)

		cpuTime := strconv.FormatUint(metric, 10)
		return &plugin.PluginMetricType{
			Namespace_: ns,
			Data_:      cpuTime,
			Timestamp_: time.Now(),
		}, nil
	}
	return nil, fmt.Errorf("Unknown error processing %v", ns)

}

func getVcpuTime(nr int, info libvirt.VirDomainInfo, dom libvirt.VirDomain) uint64 {
	var cpuTime uint64
	vcpus, err := dom.GetVcpus(int32(info.GetNrVirtCpu()))
	if err != nil {
		return cpuTime
	}
	for k, v := range vcpus {
		if k == nr {
			cpuTime = v.CpuTime
		}
	}
	return cpuTime

}

func getCPUMetricTypes(dom libvirt.VirDomain, hostname string) ([]plugin.PluginMetricType, error) {
	var mts []plugin.PluginMetricType

	domainname, err := dom.GetName()
	if err != nil {
		return nil, err
	}
	info, err := dom.GetInfo()
	if err != nil {
		return nil, err
	}
	var i uint16

	for _, metric := range cpuMetricsTypes {

		for i = 0; i < info.GetNrVirtCpu(); i++ {

			mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"libvirt", hostname, domainname, "cpu", "vcpu", strconv.FormatUint(uint64(i), 10), metric}})

		}
		mts = append(mts, plugin.PluginMetricType{Namespace_: []string{"libvirt", hostname, domainname, "cpu", metric}})
	}
	return mts, nil
}
