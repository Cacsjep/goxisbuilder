package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/Cacsjep/goxis/pkg/axmanifest"
)

const VERSION = "1.4.0"

func main() {
	showHelp := flag.Bool("h", false, "Displays this help message.")
	showVersion := flag.Bool("v", false, "Displays builders version.")
	ip := flag.String("ip", "", "The IP address of the camera where the EAP application is installed.")
	manifestPath := flag.String("manifest", "manifest.json", "The path to the manifest file. Defaults to 'manifest.json'.")
	dockerFile := flag.String("dockerfile", "", "Use a custom docker file'.")
	pwd := flag.String("pwd", "", "The root password for the camera where the EAP application is installed.")
	arch := flag.String("arch", "aarch64", "The architecture for the ACAP application: 'aarch64' or 'armv7hf'.")
	doStart := flag.Bool("start", false, "Set to true to start the application after installation.")
	doInstall := flag.Bool("install", false, "Set to true to install the application on the camera.")
	lowestSdkVersion := flag.Bool("lowsdk", false, "Set to true to build for firmware versions greater than 10.9 with SDK version 1.1. This adjusts the manifest to use version 1.3.")
	watch := flag.Bool("watch", false, "Set to true to monitor the package log after building.")
	appDirectory := flag.String("appdir", "", "The full path to the application directory from which to build.")
	filesToAdd := flag.String("files", "", "Add additional files to the container. (filename1 filename2 ...), files need to be in appdir")
	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(1)
	}

	if *showVersion {
		fmt.Println("Version:", VERSION)
		os.Exit(1)
	}

	ctx := context.Background()
	cli, err := newDockerClient()
	if err != nil {
		handleError("Failed create new docker client", err)
	}

	if *appDirectory == "" {
		fmt.Println("-appdir", "is required")
		os.Exit(1)
	}

	manifestPathFull := path.Join(*appDirectory, *manifestPath)
	amf, err := axmanifest.LoadManifest(manifestPathFull)
	if err != nil {
		handleError(fmt.Sprintf("Failed to load manifest from %s", manifestPathFull), err)
	}

	buildConfig := BuildConfiguration{
		AppDirectory: *appDirectory,
		Arch:         *arch,
		Manifest:     amf,
		ManifestPath: *manifestPath,
		Ip:           *ip,
		Pwd:          *pwd,
		DoStart:      *doStart,
		DoInstall:    *doInstall,
		LowestSdk:    *lowestSdkVersion,
		Watch:        *watch,
		Dockerfile:   *dockerFile,
		FilesToAdd:   *filesToAdd,
		ImageName:    fmt.Sprintf("%s:%s", amf.ACAPPackageConf.Setup.AppName),
	}
	// Configure SDK and architecture for the specific app
	configureSdk(*lowestSdkVersion, &buildConfig)
	configureArchitecture(*arch, &buildConfig)

	if err := buildAndRunContainer(ctx, cli, &buildConfig); err != nil {
		handleError("Failed to build and run container", err)
	}

	printCompatibility(&buildConfig)
	listEapDirectory()
	os.Remove("docker-build.log")

	if buildConfig.Watch {
		watchPackageLog(&buildConfig)
	}

}
