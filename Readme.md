# Goxisbuilder

Goxisbuilder is a CLI helper that packages Go applications into Docker-based ACAP deployments. It wraps the [`goxis`](https://github.com/Cacsjep/goxis) toolchain with sensible defaults, automatic Makefile generation, and helpers for installing/running packages on Axis cameras.

## Install

Install the latest release with `go install`:

```sh
go install github.com/Cacsjep/goxisbuilder@latest
```

On macOS, make sure your Go bin directory is on `PATH` before running the command:

```sh
export PATH="$PATH:$HOME/go/bin"
```

## Quick start (new project)

Use the bundled template to create a skeleton project:

```sh
goxisbuilder.exe -newapp
```

It will drop a single application directory with a basic `manifest.json`, a working Go module, and a `LICENSE` file so you can jump straight into development.

## Building applications

After you have an application, there are two ways to run a build:

### Inside the application directory

```sh
cd myacap
goxisbuilder.exe
```

The command expects the working directory to contain:

- `go.mod` / `go.sum`
- One or more `.go` sources (`app.go`, `main.go`, etc.)
- `manifest.json` and `LICENSE`

On success, a `build/` directory with the generated `.eap` file is created alongside your source.

### Outside (multi-application workspace)

```sh
goxisbuilder.exe -appdir=myproject
```

Each subdirectory under `myproject` that follows the single-app layout described above will be built and packaged. This is useful for monorepos or workspaces that keep multiple ACAPs together.

## Common workflows

| Flag         | Description |
|--------------|-------------|
| `-h`         | Show help. |
| `-appdir`    | Path to the application directory when invoking from a parent workspace. |
| `-arch`      | Target architecture (`aarch64` or `armv7hf`; defaults to `aarch64`). |
| `-dockerfile`| Provide a custom Dockerfile (should derive from this repo's template). |
| `-files`     | Space- or comma-separated files/directories to bundle in the final `.eap`. |
| `-install`   | Install the package on the camera after building (requires `-ip`/`-pwd`). |
| `-nocopy`    | Skip copying the resulting `.eap` file back to the host. |
| `-ip` / `-pwd` | IP address and root password for installation/start/watch commands. |
| `-lowsdk`    | Use older ACAP SDK (v3.5 on Ubuntu 20.04). |
| `-manifest`  | Path to the manifest (defaults to `manifest.json`). |
| `-newapp`    | Generate a new application scaffold. |
| `-prune`     | Run `docker system prune -f` after the build completes. |
| `-start`     | Start the installed package on the camera. |
| `-sdk`       | Specify the SDK version, e.g., `-sdk=12.2.0`. |
| `-watch`     | Tail the app log on the camera after installing. |
| `-tags`      | Go build tags forwarded through Docker/Makefile (space/comma separated). |
| `-upx`       | Enable compression of the Go binary with UPX (`true` by default). |

## Optional helpers

- **Install + start + watch**: Combine `-install -start -watch` with `-ip`/`-pwd` to deploy the build to a camera and stream its log via syslog.
- **Additional assets**: `-files` can point to model weights, configuration, or other assets that should be bundled inside the `.eap`. These paths must live in the application directory.
- **Custom Dockerfile**: Pass `-dockerfile` to override the internal Docker template. The custom file should mimic the Dockerfile in this repository.
- **Multiple manifest files**: Use `-manifest=path/to/alternate.json` when more than one manifest exists for the same app.
- **Go tags**: `-tags="prod netcgo"` becomes `GO_BUILD_TAGS` in the generated Makefile and is normalized to a comma-separated list.
- **No-copy deployments**: Add `-nocopy` when you only need to install/start/watch the application on the camera and do not care about retaining the `.eap` locally; it keeps the build artifacts inside `build/` on the container instead of copying them back to your drive.

## Custom SDK, OS, and architecture targets

Pass `-sdk`, `-arch`, and `-ubunutu` (sic) to target a particular Axis OS version and runtime. Include `-manifest` if your app ships multiple manifests, plus `-ignore` to keep large directories (such as `.git`) out of the Docker context.

### Axis OS 11.11 example

```sh
goxisbuilder.exe -appdir "./ax_msf" \
  -install -ip 10.0.0.48 -pwd 1qay2wsx \
  -sdk "1.15" -ubunutu "22.04" \
  -manifest "manifestv11.json" \
  -tags "prod" -arch aarch64 \
  -ignore ".git web website"
```

### Axis OS 12.5 example

```sh
goxisbuilder.exe -appdir "./ax_msf" \
  -install -ip 10.0.0.48 -pwd 1qay2wsx \
  -sdk "12.5.0" -watch \
  -tags "prod" -arch aarch64 \
  -ignore ".git web website"
```

The `-ignore` flag accepts space-separated values and behaves like the `_` prefix in the application directory: matching paths are excluded from the Docker build context, so the ones listed above (especially version control directories) are never copied into the container.

## Build behavior you should know

- **UPX compression**: The Docker image installs `upx-ucl` (see [Dockerfile](Dockerfile)) and compresses the Go binary with `upx --best --lzma` by default. You can disable it per build with `-upx=false`.
- **Ignored files**: Prefix a file or directory name with `_` to keep it out of the Docker context. This prevents large git history (e.g., `.git/`) or other build artifacts from being copied into the container. The builder never copies files that begin with `_`.
- **Build artifacts**: The `build/` directory is always recreated alongside your source and holds the `.eap`. Use `-nocopy` if you do not want to copy the `.eap` back to the host volume, for example when building solely to install on a camera.
- **Docker pruning**: `-prune` removes dangling Docker data after the build, which keeps disk usage down but adds runtime to the command.

## Usage reminders

```sh
.\goxisbuilder.exe -h
```

Use the generated help output for a quick flag reference if you forget a parameter name.

## Further reading

- `generate_makefile.py` - shows how the Makefile gets generated for each build.
- `Dockerfile` - contains the runtime stack, UPX installation, and environment variables that get baked into the build container.
