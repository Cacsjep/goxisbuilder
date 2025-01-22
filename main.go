package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/Cacsjep/goxis/pkg/axmanifest"
)

func main() {
	showHelp := flag.Bool("h", false, "Displays this help message.")
	ip := flag.String("ip", "", "The IP address of the camera where the EAP application is installed.")
	manifestPath := flag.String("manifest", "manifest.json", "The path to the manifest file. Defaults to 'manifest.json'.")
	dockerFile := flag.String("dockerfile", "", "Use a custom docker file'.")
	pwd := flag.String("pwd", "", "The root password for the camera where the EAP application is installed.")
	arch := flag.String("arch", "aarch64", "The architecture for the ACAP application: 'aarch64' or 'armv7hf'.")
	doStart := flag.Bool("start", false, "Set to true to start the application after installation.")
	createProject := flag.Bool("newapp", false, "Generate a new goxis app.")
	sdk_version := flag.String("sdk", "", "The version of the SDK to use. (blank = 12.2.0)")
	ubunutu_version := flag.String("ubunutu", "", "The Ubunut version to use. (blank = 24.04)")
	doInstall := flag.Bool("install", false, "Set to true to install the application on the camera.")
	prune := flag.Bool("prune", false, "Set to true execute 'docker system prune -f' after build.")
	lowestSdkVersion := flag.Bool("lowsdk", false, "Set to true to build with acap-sdk version 3.5 and ubunutu 20.04")
	watch := flag.Bool("watch", false, "Set to true to monitor the package log after building.")
	appDirectory := flag.String("appdir", "", "The path to the application directory from which to build, or blank if the current directory is the application directory.")
	filesToAdd := flag.String("files", "", "Add additional files to the container. (filename1 filename2 directory ...), files need to be in appdir")
	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(1)
	}

	if *createProject {
		createNewProject()
	}

	ctx := context.Background()
	cli, err := newDockerClient()
	if err != nil {
		handleError("Failed create new docker client", err)
	}

	if *appDirectory == "" {
		if _, err := os.Stat("go.mod"); errors.Is(err, os.ErrNotExist) {
			fmt.Println("A go project (go.mod) was not found in the current directory. Or create it (go mod init <module-path>) if you are inside the project directory.")
			os.Exit(1)
		}
		if _, err := os.Stat("LICENSE"); errors.Is(err, os.ErrNotExist) {
			fmt.Println("A LICENSE file was not found in the current directory. Please specify the app directory with -appdir, or create it if you are inside the project directory.")
			os.Exit(1)
		}
		if _, err := os.Stat("manifest.json"); errors.Is(err, os.ErrNotExist) {
			fmt.Println("A manifest.json file was not found in the current directory. Please specify the app directory with -appdir, or create it if you are inside the project directory.")
			os.Exit(1)
		}
		files, err := filepath.Glob("*.go")
		if err != nil {
			fmt.Println("Failed to search for Go files:", err)
			os.Exit(1)
		}
		if len(files) == 0 {
			fmt.Println("No Go (.go) files found in the current directory. Please specify the app directory with -appdir, or create it if you are inside the project directory.")
			os.Exit(1)
		}
	}

	manifestPathFull := path.Join(*appDirectory, *manifestPath)
	amf, err := axmanifest.LoadManifest(manifestPathFull)
	if err != nil {
		handleError(fmt.Sprintf("Failed to load manifest from %s", manifestPathFull), err)
	}

	buildConfig := BuildConfiguration{
		AppDirectory:  *appDirectory,
		Arch:          *arch,
		Manifest:      amf,
		ManifestPath:  *manifestPath,
		Ip:            *ip,
		Pwd:           *pwd,
		DoStart:       *doStart,
		DoInstall:     *doInstall,
		LowestSdk:     *lowestSdkVersion,
		Watch:         *watch,
		Dockerfile:    *dockerFile,
		FilesToAdd:    *filesToAdd,
		Prune:         *prune,
		ImageName:     fmt.Sprintf("%s:%s", *arch, amf.ACAPPackageConf.Setup.AppName),
		SdkVersion:    *sdk_version,
		UbunutVersion: *ubunutu_version,
	}
	// Configure SDK and architecture for the specific app
	configureSdk(*lowestSdkVersion, &buildConfig)
	configureArchitecture(*arch, &buildConfig)

	if err := buildAndRunContainer(ctx, cli, &buildConfig); err != nil {
		handleError("Failed to build and run container", err)
	}

	printCompatibility(&buildConfig)
	listEapDirectory()

	if buildConfig.Watch {
		watchPackageLog(&buildConfig)
	}

}
