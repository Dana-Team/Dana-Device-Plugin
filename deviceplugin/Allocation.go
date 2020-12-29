package deviceplugin

import (
	"flag"
	"fmt"
	"github.com/Dana-Team/Dana-Device-Plugin/third_party/gpuallocator"
	"golang.org/x/net/context"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"strings"
)

var passDeviceSpecs = flag.Bool("pass-device-specs", false, "pass the list of DeviceSpecs to the kubelet on Allocate()")


func (m *DanaDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	fmt.Print("start allocation\n\n")
	responses := pluginapi.AllocateResponse{}

	for _, req := range reqs.ContainerRequests {
		fmt.Print(("reqets ======= %v\n\n"),reqs.ContainerRequests)
		for _, id := range req.DevicesIDs {
			fmt.Print("id %vfdfdfdfdfdf\n\n ", id )

			if !m.DeviceExists(id) {
				return nil, fmt.Errorf("invalid allocation request for '%s': unknown device: %s", m.resourceName, id)
			}
		}
			
		response := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				m.allocateEnvvar: strings.Join(req.DevicesIDs, ","),
			},
		}
		if *passDeviceSpecs {
			response.Devices = m.ApiDeviceSpecs(req.DevicesIDs)
		}

		responses.ContainerResponses = append(responses.ContainerResponses, &response)
		fmt.Print("\n\n\n responsss ===  %v",responses.ContainerResponses	)

	}
	fmt.Print("\n\n reponse_return  %v", &responses )



	return &responses, nil
}


func (m *DanaDevicePlugin) GetPreferredAllocation(ctx context.Context, r *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	fmt.Print( " \n You Entered the GetPreferredAllocation FUNCTION \n")
	response := &pluginapi.PreferredAllocationResponse{}
	for _, req := range r.ContainerRequests {
		fmt.Print("\n req :  ",req ,"/n")
		available, err := gpuallocator.NewDevicesFrom(req.AvailableDeviceIDs)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve list of available devices: %v", err)
		}

		required, err := gpuallocator.NewDevicesFrom(req.MustIncludeDeviceIDs)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve list of required devices: %v", err)
		}

		allocated := m.allocatePolicy.Allocate(available, required, int(req.AllocationSize))

		var deviceIds []string
		for _, device := range allocated {
			deviceIds = append(deviceIds, device.UUID)
		}

		resp := &pluginapi.ContainerPreferredAllocationResponse{
			DeviceIDs: deviceIds,
		}

		response.ContainerResponses = append(response.ContainerResponses, resp)
	}
	return response, nil
}
