package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
)

func Log(v ...interface{}) {
    fmt.Println(v...)
}

var (
	prestart = flag.Bool("prestart", false, "run the prestart hook")
)

// TODO: 采用两步法来更新GPU slot
//   1. 在doPrestart()中发送资源请求，server在相应GPU slot设置占用位，如"IN"
//   2. 在exit()中，发送容器restart成功信息，server更新相应GPU标记位，为CID
func exit() {
        // if err occurred with Panic, then return error code none zero
	if err := recover(); err != nil {
		if _, ok := err.(runtime.Error); ok {
			log.Println(err)
		}
		if os.Getenv("NV_DEBUG") != "" {
			log.Printf("%s", debug.Stack())
		}
		os.Exit(1)
	}
        // normally exit
	os.Exit(0)
}

func capabilityToCLI(cap string) string {
	switch cap {
	case "compute":
		return "--compute"
	case "compat32":
		return "--compat32"
	case "graphics":
		return "--graphics"
	case "utility":
		return "--utility"
	case "video":
		return "--video"
	default:
		log.Panicln("unknown driver capability:", cap)
	}
	return ""
}

func doPrestart() {
	defer exit()
	log.SetFlags(0)

	cli := getCLIConfig()
	container := getContainerConfig()

	nvidia := container.Nvidia
	if nvidia == nil {
		// Not a GPU container, nothing to do.
		return
	}

	args := []string{cli.Path}
	if cli.LoadKmods {
		args = append(args, "--load-kmods")
	}
	if cli.Debug != nil {
		args = append(args, fmt.Sprintf("--debug=%s", *cli.Debug))
	}
	args = append(args, "configure")

	if cli.Configure.Ldconfig != nil {
		args = append(args, fmt.Sprintf("--ldconfig=%s", *cli.Configure.Ldconfig))
	}

        // 'nvidia.Devices' this variable is equal to the value of 'NVIDIA_VISIBILE_DEVICES'
        // passed through 'libcontainer-cli'. The IF condition ensure the services or containers
        // won't request a GPU from server in default, unless user has specified
        // 'NVIDIA_VISIBILE_DEVICES', but only one GPU index will be fetched.
	if len(nvidia.Devices) > 0 {
                if _, ok := container.Env[envGPUServer]; !ok {
                        log.Panicln("Before start a GPU container, please specify a correct server address, like [IP:PORT]")
                }
                if ok := checkServerAddress(container.Env[envGPUServer]); !ok {
                        log.Panicln("Bad server address ", container.Env[envGPUServer])
                }

                node, task, index := requestResource(container.Env[envGPUServer], container.Env["HOSTNAME"])
                log.Printf("Recived GPU info : %s : %s : %d", node, task, index)

                nvidia.Devices = index
                container.Env[envNVGPU] = index
                log.SetOutput(os.Stdout)
                log.Printf("nvidia.Devices = %s, container.Env = %s", nvidia.Devices, container.Env[envNVGPU])

		args = append(args, fmt.Sprintf("--device=%s", index))
	}


	for _, cap := range strings.Split(nvidia.Capabilities, ",") {
		if len(cap) == 0 {
			break
		}
		args = append(args, capabilityToCLI(cap))
	}

	if !cli.DisableRequire && !nvidia.DisableRequire {
		for _, req := range nvidia.Requirements {
			args = append(args, fmt.Sprintf("--require=%s", req))
		}
	}

	args = append(args, fmt.Sprintf("--pid=%s", strconv.FormatUint(uint64(container.Pid), 10)))
	args = append(args, container.Rootfs)

	env := append(os.Environ(), cli.Environment...)

        // Output basic info
        log.SetOutput(os.Stderr)
	log.Printf("this is my test printing.....exec command: %v || and env.... %s \n", args, env)

	err := syscall.Exec(cli.Path, args, env)
	log.Panicln("exec failed:", err)
}

func main() {
	flag.Parse()

	if *prestart {
		doPrestart()
	}
}
