package deviceplugin

import (
	"github.com/Dana-Team/Dana-Device-Plugin/third_party/gpuallocator"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"golang.org/x/net/context"
)

// DanaDevicePlugin implements the Kubernetes device plugin API
type DanaDevicePlugin struct {
	ResourceManager
	resourceName   string
	socket string
	allocatePolicy gpuallocator.Policy
	allocateEnvvar string

	server *grpc.Server
	cachedDevices []*Device
	avalDevices []*Device
	health chan *Device
	stop   chan interface{}
}
type allocatePolicy gpuallocator.Policy

func NewDanaDevicePlugin(resourceName string, resourceManager ResourceManager,  allocateEnvvar string, socket string) *DanaDevicePlugin {
	return &DanaDevicePlugin{
		ResourceManager: resourceManager,
		resourceName:    resourceName,
		socket:          socket,
		allocateEnvvar:	allocateEnvvar,
		avalDevices: nil,

		// These will be reinitialized every
		// time the plugin server is restarted.
		cachedDevices: nil,
		server:        nil,
		health:        nil,
		stop:          nil,
	}
}


func (m *DanaDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}


func (m *DanaDevicePlugin) DeviceExists(id string) bool {
	for _, d := range m.cachedDevices {
		if d.ID == id {
			return true
		}
	}
	return false
}

func (m *DanaDevicePlugin) ApiDevices() []*pluginapi.Device {
	var pdevs []*pluginapi.Device
	for _, d := range m.cachedDevices {
		pdevs = append(pdevs, &d.Device)
	}
	return pdevs
}

func (m *DanaDevicePlugin) ApiDeviceSpecs(filter []string) []*pluginapi.DeviceSpec {
	var specs []*pluginapi.DeviceSpec

	paths := []string{
		"/run/nvidia/driver/dev/nvidiactl",
		"/run/nvidia/driver/dev/nvidia-uvm",
		"/run/nvidia/driver/dev/nvidia-uvm-tools",
		"/run/nvidia/driver/dev/nvidia-modeset",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			spec := &pluginapi.DeviceSpec{
				ContainerPath: p,
				HostPath:      p,
				Permissions:   "rw",
			}
			specs = append(specs, spec)
		}
	}

	for _, d := range m.cachedDevices {
		for _, id := range filter {
			if d.ID == id {
				spec := &pluginapi.DeviceSpec{
					ContainerPath: d.Path,
					HostPath:      d.Path,
					Permissions:   "rw",
				}
				specs = append(specs, spec)
			}
		}
	}

	return specs
}
