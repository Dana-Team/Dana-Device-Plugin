package deviceplugin

import (
	"flag"
	"fmt"
	"github.com/Dana-Team/Dana-Device-Plugin/third_party/gpuallocator"
	"golang.org/x/net/context"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"strings"
	"unicode/utf8"

	//"strconv"
)

var passDeviceSpecs = flag.Bool("pass-device-specs", false, "pass the list of DeviceSpecs to the kubelet on Allocate()")


func (m *DanaDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	fmt.Print("start allocation\n\n")
	responses := pluginapi.AllocateResponse{}

	var reqs2 *pluginapi.AllocateRequest
	reqs2 = reqs
	j := 0
	for _, req := range reqs.ContainerRequests {
		fmt.Print(("reqets  :\n\n"),reqs.ContainerRequests)
		i := 0
		for _, id := range req.DevicesIDs {

			fmt.Print("\n id    :", id, "\n" )

			s:=trimLastChar(id)

			fmt.Print("\n id after remove  :",s,"\n")

			reqs2.ContainerRequests[j].DevicesIDs[i] = s
			i++

			if !m.DeviceExists(s) {
				return nil, fmt.Errorf("invalid allocation request for '%s': unknown device: %s", m.resourceName, id)
			}

		}

		fmt.Print("\n req2    : ",reqs2,"\n")
		response := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				m.allocateEnvvar: strings.Join(reqs2.ContainerRequests[j].DevicesIDs, ","),
			},
		}
		if *passDeviceSpecs {
			response.Devices = m.ApiDeviceSpecs(reqs2.ContainerRequests[j].DevicesIDs)
		}

		responses.ContainerResponses = append(responses.ContainerResponses, &response)
		fmt.Print("\n\n\n responsss ===  %v",responses.ContainerResponses	)
		j++
	}
	fmt.Print("\n\n reponse_return  %v", &responses )



	return &responses, nil
}



func (m *DanaDevicePlugin) GetPreferredAllocation(ctx context.Context, r *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	fmt.Print( " \n You Entered the GetPreferredAllocation FUNCTION \n")
	response := &pluginapi.PreferredAllocationResponse{}


	for _, req := range r.ContainerRequests {

		for j, i := range req.AvailableDeviceIDs {
			fmt.Print("\n",i,"\n")
			s:=trimLastChar(i)

			fmt.Print("\nafter trim ",s,"\n")
			req.AvailableDeviceIDs[j] = s
		}



		available, err := gpuallocator.NewDevicesFrom(req.AvailableDeviceIDs)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve list of available devices: %v", err)
		}
		required, err := gpuallocator.NewDevicesFrom(req.MustIncludeDeviceIDs)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve list of required devices: %v", err)
		}


		if req.AllocationSize != 1 {
			fmt.Print("\n Asked for more then 1 device",req.AllocationSize,"\n")
		}
		allocated := m.allocatePolicy.Allocate(available, required, int(req.AllocationSize))

		fmt.Print("\nAllocated  - ",allocated,"\n")

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

func trimLastChar(s string) string {
	r, size := utf8.DecodeLastRuneInString(s)
	if r == utf8.RuneError && (size == 0 || size == 1) {
		size = 0
	}
	return s[:len(s)-size-1]
}