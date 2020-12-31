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

	for _, req := range reqs.ContainerRequests {
		fmt.Print(("reqets  :\n\n"),reqs.ContainerRequests)
		for _, id := range req.DevicesIDs {

			fmt.Print("\n id    :", id, "\n" )

			//realid :=strings.Trim(id,"fake")
			//s := id
			//sz := len(s)
		//	if sz > 0 && s[sz-1]== '+' {
		//		s=s[:sz-1]
		//	}
//
		//	id = s
			fmt.Print("\n id after remove  :",id,"\n")


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
	 fmt.Print("\n AvailableDeviceIDs 1  :%s" ,r.ContainerRequests[0].AvailableDeviceIDs[0],"\n")
	fmt.Print("\n AvailableDeviceIDs  2 :%s" ,r.ContainerRequests[0].AvailableDeviceIDs[1],"\n")
	fmt.Print("\n AvailableDeviceIDs  3 :%s" ,r.ContainerRequests[0].AvailableDeviceIDs[2],"\n")
	fmt.Print("\n AvailableDeviceIDs  4 :%s" ,r.ContainerRequests[0].AvailableDeviceIDs[3],"\n")
	fmt.Print("\n AvailableDeviceIDs  ALL :%s" ,r.ContainerRequests[0],"\n")

	for _, req := range r.ContainerRequests {

		for j, i := range req.AvailableDeviceIDs {
			fmt.Print("\n",i,"\n")
			s:=trimLastChar(i)

			fmt.Print("\nafter trim ",s,"\n")
			req.AvailableDeviceIDs[j] = s
		}


		fmt.Print("\n before AvailableDeviceIDs \n")

		available, err := gpuallocator.NewDevicesFrom(req.AvailableDeviceIDs)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve list of available devices: %v", err)
		}
		fmt.Print("\n passed AvailableDeviceIDs \n")
		required, err := gpuallocator.NewDevicesFrom(req.MustIncludeDeviceIDs)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve list of required devices: %v", err)
		}
		fmt.Print("\n passed MustIncludeDeviceIDs \n")

		allocated := m.allocatePolicy.Allocate(available, required, int(req.AllocationSize))
		fmt.Print("\n passed allocated \n")

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
	return s[:len(s)-size]
}