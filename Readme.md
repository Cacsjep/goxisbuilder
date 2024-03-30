## Goxisbuilder

Goxisbuilder is a powerful command-line tool designed to streamline the process of building Docker ACAP applications for Go developers. 

```shell
go install github.com/Cacsjep/goxisbuilder@latest
```

### Application Structure Requirements
Before building an ACAP application with Goxisbuilder, you need to prepare your application directory. This directory should contain all the necessary components for your application to be successfully built:

- **Application Directory**: Your application must reside in its own directory. This directory will contain the files listed below and is referred to by the `-appdir` flag during the build process.
  - **.go File**: A Go source file with a `main` function. This file acts as the entry point of your application.
  - **LICENSE**: A file that specifies the licensing terms under which your application is distributed. It's crucial to include this to inform users of how they can legally use your application.
  - **manifest.json**: A JSON file that provides detailed information about your application, including its name, version, and any dependencies it might have.

Ensure that your application directory is properly structured and contains these components before proceeding with the build process.

### Important Notes
> The -appdir flag is required. It specifies the directory of the application you wish to build. This directory must contain a main.go file, a LICENSE file, and a manifest.json file.

> Ensure that the LICENSE and manifest.json files are correctly formatted and contain all necessary information as per your application's requirements.

### Usage

```shell
.\goxisbuilder.exe -appdir="mycoolacap"
```

| Flag              | Description                                                                                                                      | Default           |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------- | ----------------- |
| `-h`              | Displays this help message.                                                                                                      |                   |
| `-appdir`         | The full path to the application directory from which to build.                                                                  | required          |
| `-arch`           | The architecture for the ACAP application: 'aarch64' or 'armv7hf'.                                                               | `"aarch64"`       |
| `-build-examples` | Set to true to build example applications.                                                                                       | `false`           |
| `-install`        | Set to true to install the application on the camera.                                                                            | `false`           |
| `-libav`          | Set to true to compile libav for binding with go-astiav.                                                                         | `false`           |
| `-lowsdk`         | Set to true to build for firmware versions greater than 10.9 with SDK version 1.1. This adjusts the manifest to use version 1.3. | `false`           |
| `-manifest`       | The path to the manifest file. Defaults to 'manifest.json'.                                                                      | `"manifest.json"` |
| `-ip`             | The IP address of the camera where the EAP application is installed.                                                             | `""`              |
| `-pwd`            | The root password for the camera where the EAP application is installed.                                                         | `""`              |
| `-start`          | Set to true to start the application after installation.                                                                         | `false`           |
| `-watch`          | Set to true to monitor the package log after building.                                                                           | `false`           |
