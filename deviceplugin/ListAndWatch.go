package deviceplugin

import (
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"log"
	"fmt"
)

// ListAndWatch lists devices and update that list according to the health status
func (m *DanaDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	s.Send(&pluginapi.ListAndWatchResponse{Devices: m.ApiDevices()})
	fmt.Print((&pluginapi.ListAndWatchResponse{Devices: m.ApiDevices()}))
	fmt.Print("THIS iS LISTANDWATCH 1 \n\n")
	for {
		select {
		case <-m.stop:
			return nil
		case d := <-m.health:
			d.Health = pluginapi.Unhealthy
			log.Printf("'%s' device marked unhealthy: %s", m.resourceName, d.ID)
			s.Send(&pluginapi.ListAndWatchResponse{Devices: m.ApiDevices()})
			fmt.Print((&pluginapi.ListAndWatchResponse{Devices: m.ApiDevices()}))
			fmt.Print("THIS iS LISTANDWATCH 2 \n\n")

		}
	}
}