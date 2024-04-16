package main

import (
	"archive/tar"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/vbauerster/mpb"
)

// newDockerClient initializes a new Docker client
func newDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return cli, nil
}

// buildAndRunContainer builds a Docker image and runs a container from it
func buildAndRunContainer(ctx context.Context, cli *client.Client, bc *BuildConfiguration) error {
	// Build Docker image
	if err := dockerBuild(ctx, cli, bc); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	// Create and start container
	containerID, err := createContainer(ctx, cli, bc.ImageName)
	if err != nil {
		return fmt.Errorf("create container failed: %w", err)
	}

	if err := copyFromContainer(ctx, cli, containerID); err != nil {
		return fmt.Errorf("copy eap failed: %w", err)
	}

	if err := cli.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		panic(err)
	}

	if err := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{}); err != nil {
		panic(err)
	}

	return nil
}

// dockerBuild performs the Docker image build operation and processes the output
func dockerBuild(ctx context.Context, cli *client.Client, bc *BuildConfiguration) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to create current dir: %w", err)
	}

	buildContext, err := createBuildContext(currentDir, bc.Dockerfile)
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}

	files_to_add := ""
	if bc.FilesToAdd != "" {
		files := strings.Split(bc.FilesToAdd, " ")
		for _, file := range files {
			files_to_add += fmt.Sprintf("-a %s ", file)
		}
	}

	options := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{bc.ImageName},
		BuildArgs: map[string]*string{
			"ARCH":                 ptr(bc.Arch),
			"SDK":                  ptr(bc.Sdk),
			"UBUNTU_VERSION":       ptr(bc.UbunutVersion),
			"VERSION":              ptr(bc.Version),
			"GO_ARCH":              ptr(bc.GoArch),
			"GO_ARM":               ptr(bc.GoArm),
			"APP_NAME":             ptr(bc.Manifest.ACAPPackageConf.Setup.AppName),
			"APP_MANIFEST":         ptr(bc.ManifestPath),
			"IP_ADDR":              ptr(bc.Ip),
			"PASSWORD":             ptr(bc.Pwd),
			"START":                ptr(boolToStr(bc.DoStart)),
			"INSTALL":              ptr(boolToStr(bc.DoInstall)),
			"GO_APP":               ptr(bc.AppDirectory),
			"CROSS_PREFIX":         ptr(bc.CrossPrefix),
			"COMP_LIBAV":           ptr(boolToStr(bc.WithLibav)),
			"FILES_TO_ADD_TO_ACAP": ptr(files_to_add),
		},
		Remove: true,
	}

	buildResponse, err := cli.ImageBuild(ctx, buildContext, options)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer buildResponse.Body.Close()
	decoder := json.NewDecoder(buildResponse.Body)

	stepRegexp := regexp.MustCompile(`Step (\d+)/(\d+)`)
	errorRegexp := regexp.MustCompile(`(?i)(error|failed|Illegal|cannot|could not|can't|\bfail\b|panic:|undefined|missing|expected|unexpected|cannot find package|no package found)`)
	p := mpb.New(mpb.WithWidth(60))
	var bar *ProgressBar
	currentStep, totalSteps := 0, 0
	errorsDetected := false
	var errorMessages []string

	f, err := os.OpenFile("docker-build.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open build log: %s", err.Error())
	}
	defer f.Close()
	log.SetOutput(f)

	for {
		var m map[string]interface{}
		if err := decoder.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("failed to decode build response: %s", err.Error())
		}

		if stream, ok := m["stream"].(string); ok {

			log.Print(stream)
			if bar != nil {
				bar.Render()
			}

			if errorsDetected {
				// Accumulate error messages after an error has been detected
				errorMessages = append(errorMessages, stream)
				continue // Skip further processing in the error state
			}

			// Check for error patterns in the stream message
			if errorRegexp.MatchString(stream) {
				errorsDetected = true
				errorMessages = append(errorMessages, stream) // Capture the first error message
				if bar != nil {
					bar.Complete()
				}
				continue
			}

			matches := stepRegexp.FindStringSubmatch(stream)
			if len(matches) == 3 {
				newTotal, _ := strconv.Atoi(matches[2])
				if newTotal > totalSteps {
					totalSteps = newTotal
					if bar == nil {
						steps := totalSteps
						if bc.DoInstall {
							steps++
						}
						if bc.DoStart {
							steps++
						}
						if bc.Watch {
							steps++
						}
						bar = &ProgressBar{current: 0, total: steps, prefix: "Starting...", spinner: []string{"-", "/", "|", "\\"}}
						bar.StartSpinner()

					} else {
						bar.SetTotal(totalSteps)
					}
				}

				if currentStep < totalSteps {
					bar.Increment()
					currentStep++
				}
			}

			if strings.Contains(stream, "ARG ARCH") {
				bar.SetPrefix("Docker prepare image...")
			}

			if strings.Contains(stream, "RUN apt-get update") {
				bar.SetPrefix("Install packages (apt)...")
			}

			if strings.Contains(stream, "ARG GOLANG_VERSION") {
				bar.SetPrefix("Install golang ...")
			}

			if strings.Contains(stream, "Building FFmpeg") {
				bar.SetPrefix("Build libav ...")
			}

			if strings.Contains(stream, "RUN python ge") {
				bar.SetPrefix("Generate Makefile ...")
			}

			if strings.Contains(stream, "RUN . /opt/axis/acapsdk/environment-setup") {
				bar.SetPrefix("Prepare application...")
			}

			if strings.Contains(stream, "go build -ldfl") {
				bar.SetPrefix("Building application...")
			}

			if strings.Contains(stream, "Create pack") {
				bar.SetPrefix("Create package...")
			}

			if strings.Contains(stream, "installing") {
				bar.SetPrefix("Install package...")
				bar.Increment()
			}

			if strings.Contains(stream, "starting") {
				bar.SetPrefix("Starting package...")
				bar.Increment()
			}

			if strings.Contains(stream, "RUN mv *.eap /opt/eap") {
				bar.SetPrefix("Copy package...")
			}
		}
	}

	p.Wait() // Wait for all bars to complete

	if errorsDetected {
		fmt.Println("\nBuild errors detected:")
		for _, errMsg := range errorMessages {
			fmt.Println(errMsg)
		}
		return errors.New("error detected during build process")
	}

	return nil
}

// createContainer creates and starts a Docker container from an image
func createContainer(ctx context.Context, cli *client.Client, imageName string) (string, error) {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
	}, nil, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("container creation failed: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("container start failed: %w", err)
	}

	return resp.ID, nil
}

// copyFromContainer copy our build result
func copyFromContainer(ctx context.Context, cli *client.Client, id string) error {
	copyFromContainer, _, err := cli.CopyFromContainer(ctx, id, "/opt/eap")
	if err != nil {
		return err
	}
	defer copyFromContainer.Close()

	os.Mkdir("eap", 0664)

	tr := tar.NewReader(copyFromContainer)
	var foundFile bool
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg {
			outputFile, err := os.Create(header.Name)
			if err != nil {
				continue
			}
			defer outputFile.Close()

			if _, err := io.Copy(outputFile, tr); err != nil {
				continue
			}
			foundFile = true
		}
	}

	if !foundFile {
		return errors.New("no file found in the archive")
	}

	return nil
}
