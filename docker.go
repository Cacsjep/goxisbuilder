package main

import (
	"archive/tar"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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

	if !bc.NotCopy {
		if err := copyFromContainer(ctx, cli, containerID); err != nil {
			return fmt.Errorf("copy eap failed: %w", err)
		}
	} else {
		fmt.Println("Copy eap file skipped")
	}

	if err := cli.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		panic(err)
	}

	if err := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{}); err != nil {
		panic(err)
	}
	if bc.Prune {
		err = exec.Command("docker", "system", "prune", "-f").Run()
		if err != nil {
			fmt.Printf("Error removing dangling images: %s\n", err)
		}
	}

	return nil
}

// dockerBuild performs the Docker image build operation and processes the output
func dockerBuild(ctx context.Context, cli *client.Client, bc *BuildConfiguration) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to create current dir: %w", err)
	}

	fmt.Println("Building Docker image...")
	buildContext, err := createBuildContext(currentDir, bc.Dockerfile, bc.IgnoreDirs)
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}

	fmt.Println("Adding files to build context...")

	files_to_add := ""
	if bc.FilesToAdd != "" {
		files := strings.Split(bc.FilesToAdd, " ")
		for _, file := range files {
			files_to_add += fmt.Sprintf("-a %s ", file)
		}
	}

	options := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{bc.ImageName, bc.Arch, bc.Sdk},
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
			"DONT_COPY":            ptr(boolToStr(bc.NotCopy)),
			"GO_APP":               ptr(bc.AppDirectory),
			"FILES_TO_ADD_TO_ACAP": ptr(files_to_add),
			"GO_BUILD_TAGS":        ptr(bc.BuildTags),
			"ENABLE_UPX":           ptr(boolToStr(bc.EnableUpx)),
		},
		Remove:      true,
		ForceRemove: true,
		NoCache:     false,
	}

	fmt.Println("Starting Docker image build...")

	buildResponse, err := cli.ImageBuild(ctx, buildContext, options)
	if err != nil {
		return fmt.Errorf("unable to build image: %w", err)
	}
	defer buildResponse.Body.Close()
	decoder := json.NewDecoder(buildResponse.Body)

	for {
		var m map[string]interface{}
		if err := decoder.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("unable to decode build response: %s", err.Error())
		}

		s, ok := m["stream"]
		if ok {
			// exit when we see the acap-build error message but when it is not a step, eg command line in dockerfile
			if strings.Contains(s.(string), "acap-build error") && !strings.Contains(s.(string), "Step") {
				return errors.New(s.(string))
			}
			fmt.Print(s)
		}
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
	copyFromContainer, _, err := cli.CopyFromContainer(ctx, id, "/opt/build")
	if err != nil {
		return err
	}
	defer copyFromContainer.Close()

	if _, err := os.Stat("build"); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir("build", os.FileMode(0755))
			if err != nil {
				return fmt.Errorf("failed to create build directory (local): %w", err)
			}
		} else {
			return fmt.Errorf("failed to check build directory (local): %w", err)
		}
	}

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
				return fmt.Errorf("failed to create file that is extracted from docker context archiv, File:%s from docker folder /opt/build, Error: %w", header.Name, err)
			}
			defer outputFile.Close()

			if _, err := io.Copy(outputFile, tr); err != nil {
				if err != nil {
					return fmt.Errorf("failed to copy file that is extracted from docker context archiv, File:%s from docker folder /opt/build, Error: %w", header.Name, err)
				}
			}
			foundFile = true
		}
	}

	if !foundFile {
		return errors.New("there is no file in the docker context archive /opt/build, but at least .eap acap file should be there")
	}

	return nil
}
