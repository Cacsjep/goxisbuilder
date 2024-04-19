## Goxisbuilder

Goxisbuilder is a powerful command-line tool designed to streamline the process of building Docker ACAP applications for Go developers. 
Its main purpose is to build apps with [goxis](https://github.com/Cacsjep/goxis).

```shell
go install github.com/Cacsjep/goxisbuilder@latest
```

[![Discord](https://img.shields.io/badge/Discord-Join%20us-blue?style=for-the-badge&logo=discord)](https://discord.gg/we6EqDSJ)

### Application Structure Requirements
Before building an ACAP application with Goxisbuilder, you need to prepare your application directory. 

This directory should contain all the necessary components for your application to be successfully built:

- **Application Directory**: Your application must reside in its own directory. This directory will contain the files listed below and is referred to by the `-appdir` flag during the build process.
  - **.go File**: A Go source file with a `main` function. This file acts as the entry point of your application.
  - **LICENSE**: A file that specifies the licensing terms under which your application is distributed. It's crucial to include this to inform users of how they can legally use your application.
  - **manifest.json**: A JSON file that provides detailed information about your application, including its name, version, and any dependencies it might have.

Ensure that your application directory is properly structured and contains these components before proceeding with the build process.

* myproject
   * go.sum
   * go.mod
     * myacap
       * *.go (app.go or main.go does not matter) 
       * manifest.json
       * LICENSE

### Important Notes
> [!IMPORTANT] 
> goxisbuilder needs to be executed in a go project, so before run goxisbuilder initalize a new project with `go mod init myproject`

> [!IMPORTANT] 
> The `-appdir` flag is required. It specifies the directory of the application you wish to build. This directory must contain a ***.go*** file with main function, a ***LICENSE*** file, and a ***manifest.json*** file.

### Usage

```shell
.\goxisbuilder.exe -appdir="mycoolacap"
```

| Flag                | Description                                                                                                                      | Default           |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------- | ----------------- |
| `-h`                | Displays this help message.                                                                                                      |                   |
| `-appdir`           | The full path to the application directory from which to build.                                                                  | required          |
| `-arch`             | The architecture for the ACAP application: 'aarch64' or 'armv7hf'.                                                               | `"aarch64"`       |
| `-dockerfile`       | Use your own dockerfile                                                                                                          | `""`       |
| `-files`            | Files for adding to the acap eap package like larod models (filename1 filename2 ...)                                             | `""`       |
| `-install`          | Set to true to install the application on the camera.                                                                            | `false`           |
| `-prune`            | Set to true execute 'docker system prune -f' after build.                                                                        | `false`           |
| `-lowsdk`           | Set to true to build with acap-sdk version 3.5 and ubunutu 20.04                                                                 | `false`           |
| `-manifest`         | The path to the manifest file. Defaults to 'manifest.json'.                                                                      | `"manifest.json"` |
| `-ip`               | The IP address of the camera where the EAP application is installed.                                                             | `""`              |
| `-pwd`              | The root password for the camera where the EAP application is installed.                                                         | `""`              |
| `-start`            | Set to true to start the application after installation.                                                                         | `false`           |
| `-watch`            | Set to true to monitor the package log after building.                                                                           | `false`           |

### Customize Docker Build
You can set your one docker file via `-dockerfile`,
just use the repo Dockerfile as starting point.

#### Example
```
goxisbuilder.exe -appdir="./mycoolacap" -dockerfile="CustomDockerfile"
```


### Additional ACAP/EAP Package files
Deploying ACAP's for example with Machine Learning models,
needs to bundle model files into the .eap package.

Just use `-files` arg to tell goxisbuilder which files you want to bundle.

> [!IMPORTANT] 
>This files need also be in **appdir**

#### Example
```
goxisbuilder.exe -appdir="./examples/axlarod/object_detection" -files=ssd_mobilenet_v2_coco_quant_postprocess.tflite
```
