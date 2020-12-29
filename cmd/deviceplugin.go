package main


import (
	"flag"
	"fmt"
	dp "github.com/Dana-Team/Dana-Device-Plugin/deviceplugin"
	nvml "github.com/Dana-Team/Dana-Device-Plugin/third_party/nvml"
	"github.com/fsnotify/fsnotify"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"log"
	"os"
	"syscall"
)

func getAllPlugins() []*dp.DanaDevicePlugin {
	return []*dp.DanaDevicePlugin{
		dp.NewDanaDevicePlugin("dana.894/gpu",
			dp.NewGpuDeviceManager(),
			"NVIDIA_VISIBLE_DEVICES" ,
			pluginapi.DevicePluginPath + "nvidia-dp2idan14.sock"),
	}
}

func main() {
	flag.Parse()

	log.Println("Loading NVML")
	log.Println("devsapce test update2check git hatul")
	if err := nvml.Init(); err != nil {
		log.Printf("Failed to initialize NVML: %s.", err)
		log.Printf("If this is a GPU node, did you set the docker default runtime to `nvidia`?")
		log.Printf("You can check the prerequisites at: https://github.com/NVIDIA/k8s-device-plugin#prerequisites")
		log.Printf("You can learn how to set the runtime at: https://github.com/NVIDIA/k8s-device-plugin#quick-start")

		select {}
	}
	defer func() { log.Println("Shutdown of NVML returned:", nvml.Shutdown()) }()

	log.Println("Starting FS watcher.")
	watcher, err := dp.NewFSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		log.Println("Failed to created FS watcher.")
		fmt.Print(err)
		os.Exit(1)
	}
	defer watcher.Close()

	log.Println("Starting OS watcher.")
	sigs := dp.NewOSWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	log.Println("Retreiving plugins.")
	plugins := getAllPlugins()

	for _, p := range plugins {
		fmt.Print(p)
		defer p.Stop()
	}


restart:
	// Loop through all plugins, idempotently stopping them, and then starting
	// them if they have any devices to serve. If even one plugin fails to
	// start properly, try starting them all again.
	started := 0
	pluginStartError := make(chan struct{})
	for _, p := range plugins {
		p.Stop()

		// Just continue if there are no devices to serve for plugin p.
		if len(p.Devices()) == 0 {
			continue
		}

		// Start the gRPC server for plugin p and connect it with the kubelet.
		if err := p.Start(); err != nil {
			log.Println("Could not contact Kubelet, retrying. Did you enable the device plugin feature gate?")
			log.Printf("You can check the prerequisites at: https://github.com/NVIDIA/k8s-device-plugin#prerequisites")
			log.Printf("You can learn how to set the runtime at: https://github.com/NVIDIA/k8s-device-plugin#quick-start")
			close(pluginStartError)
			goto events
		}
		started++
	}

	if started == 0 {
		log.Println("No devices found. Waiting indefinitely.")
	}


events:
	// Start an infinite loop, waiting for several indicators to either log
	// some messages, trigger a restart of the plugins, or exit the program.
	for {
		select {
		// If there was an error starting any plugins, restart them all.
		case <-pluginStartError:
			goto restart

		// Detect a kubelet restart by watching for a newly created
		// 'pluginapi.KubeletSocket' file. When this occurs, restart this loop,
		// restarting all of the plugins in the process.
		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				log.Printf("inotify: %s created, restarting.", pluginapi.KubeletSocket)
				goto restart
			}

		// Watch for any other fs errors and log them.
		case err := <-watcher.Errors:
			log.Printf("inotify: %s", err)

		// Watch for any signals from the OS. On SIGHUP, restart this loop,
		// restarting all of the plugins in the process. On all other
		// signals, exit the loop and exit the program.
		case s := <-sigs:
			switch s {
			case syscall.SIGHUP:
				log.Println("Received SIGHUP, restarting.")
				goto restart
			default:
				log.Printf("Received signal \"%v\", shutting down.", s)
				for _, p := range plugins {
					p.Stop()
				}
				break events
			}
		}
	}
}
