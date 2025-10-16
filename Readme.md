## Goxisbuilder

Goxisbuilder is a powerful command-line tool designed to streamline the process of building Docker ACAP applications for Go developers. 
Its main purpose is to build apps with [goxis](https://github.com/Cacsjep/goxis).

### Install
```shell
go install github.com/Cacsjep/goxisbuilder@latest
```

> [!NOTE] 
> MAC OS: If you not already have **go bin directory** in your path you need to perform `export PATH=$PATH:$HOME/go/bin`

### Quick Start (New Project)
Creating a new project is very handy, it creates an application directory with all the necessary stuff inside. :)
```shell
goxisbuilder.exe -newapp
```

[![Discord](https://img.shields.io/badge/Discord-Join%20us-blue?style=for-the-badge&logo=discord)](https://discord.gg/we6EqDSJ)

### Building Applications
There are two ways of building apps with goxisbuilder:
- Inside an application directory
- Outside an application directory"

#### Inside a Applications Directory
Build command: `goxisbuilder.exe` or on linux `goxisbuilder`

Directory/File Structure:
* myacap
   * go.sum
   * go.mod
   * *.go (app.go or main.go does not matter) 
   * manifest.json
   * LICENSE


> [!NOTE] 
> After a successful build, a **build** directory with the corresponding .eap file is created.

#### Outside a Applications Directory
Build command: `goxisbuilder.exe -appdir=<application-director>` or on linux `goxisbuilder -appdir=<application-director>`

Directory/File Structure:
* myproject
   * go.sum
   * go.mod
     * myacap1
       * *.go (It does not matter whether you use app.go or main.go.) 
       * manifest.json
       * LICENSE
     * myacap2
       * *.go (It does not matter whether you use app.go or main.go.) 
       * manifest.json
       * LICENSE

> [!NOTE] 
> After a successful build, a **build** directory with the corresponding .eap file is created.

### Start, Install, Watch
To install and start the ACAP application after building it, add the `-install`  and `-start`  flags. Also, specify the `-ip <camera IP address>`  and `-pwd <camera root password>`  flags.

If you are interested in viewing the syslog output of the ACAP application, add the `-watch`  flag. (Note: IP address and password are required.)

### Additional ACAP/EAP Package files
When deploying ACAPs, such as those with machine learning models, it's necessary to bundle model files into the .eap package.

Simply use the `-files` argument to specify which files goxisbuilder should bundle.

> [!IMPORTANT] 
> These files need to be in the Application Directory.

#### Example
```
goxisbuilder.exe -files=ssd_mobilenet_v2_coco_quant_postprocess.tflite
```

### Custom Dockerfile
To use your own Dockerfile, add the `-dockerfile` flag and base it on the repository's *Dockerfile*.

### Multiple Manifests
In case of multiple manifest files for an application, you can use the `-manifest` flag to specify which manifest file to use for the build.

### Ignore dirs or files
Files or directories can be excluded from being copied into the container by prefixing their names with an underscore.

### Usage
```shell
.\goxisbuilder.exe -h
```

| Flag                | Description                                                                                                                      | Default           |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------- | ----------------- |
| `-h`                | Displays this help message.                                                                                                      |                   |
| `-appdir`           | The path to the application directory from which to build.                                                                       | `""`              |
| `-arch`             | The architecture for the ACAP application: 'aarch64' or 'armv7hf'.                                                               | `"aarch64"`       |
| `-dockerfile`       | Use your own dockerfile                                                                                                          | `""`              |
| `-files`            | Files for adding to the acap eap package like larod models (filename1 filename2 directory)                                             | `""`              |
| `-install`          | Set to true to install the application on the camera.                                                                            | `false`           |
| `-nocopy`             | Set to true when you dont need the eap file after build.                                                                        | `false`           |
| `-ip`               | The IP address of the camera where the EAP application is installed.                                                             | `""`              |
| `-lowsdk`           | Set to true to build with acap-sdk version 3.5 and ubunutu 20.04                                                                 | `false`           |
| `-manifest`         | The path to the manifest file. Defaults to 'manifest.json'.                                                                      | `"manifest.json"` |
| `-newapp  `         | Generate a new goxis application                                                                                                 | `false`           |
| `-prune`            | Set to true execute 'docker system prune -f' after build.                                                                        | `false`           |
| `-pwd`              | The root password for the camera where the EAP application is installed.                                                         | `""`              |
| `-start`            | Set to true to start the application after installation.                                                                         | `false`           |
| `-sdk`              | Set to specifiy the sdk version for the ACAP image like, for example -version 1.12                                               | `12.2.0`           |
| `-watch`            | Set to true to monitor the package log after building.                                                                           | `false`           |
| `-tags`             | Go build tags passed through Docker to Makefile and applied to `go build -tags`. Accepts space- or comma-separated values; normalized to comma-separated.           | `""`              |

#### Build tags examples
- Single tag: `.\goxisbuilder.exe -tags=prod`
- Multiple tags (space): `.\goxisbuilder.exe -tags="prod netcgo"`
- Multiple tags (comma): `.\goxisbuilder.exe -tags=prod,netcgo`

These are forwarded to Docker as `GO_BUILD_TAGS` and used by the generated Makefile. Tags are normalized to a comma-separated list internally.



