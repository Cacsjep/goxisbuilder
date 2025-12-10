# sdk v3.5 build script has no way to specify manifest file an has also not the arg --build no-build
# so this script bypass this by renaming the given manifest to manifest.json and create the makefile for golang build

import os
import sys


def create_makefile(app_name, appdir, manifest_file_name):
    if appdir != ".":
        default_manifest = os.path.join(appdir, "manifest.json")
        custom_manifest = os.path.join(appdir, manifest_file_name)
        make_file_path = os.path.join(appdir, "Makefile")
    else:
        default_manifest = "manifest.json"
        custom_manifest = manifest_file_name
        make_file_path = "Makefile"
    if custom_manifest != default_manifest:
        if os.path.exists(default_manifest): 
            print(f"Removing existing default manifest: {default_manifest}")
            os.remove(default_manifest)
        print(f"Renaming {custom_manifest} to {default_manifest}")
        os.rename(custom_manifest, default_manifest)
        print(f"Manifest renamed from {manifest_file_name} to {default_manifest}")
        # Verify the rename worked by reading the schema version
        try:
            import json
            with open(default_manifest, 'r') as f:
                manifest_data = json.load(f)
                schema_version = manifest_data.get('schemaVersion', 'unknown')
                print(f"Verified: {default_manifest} now has schemaVersion: {schema_version}")
        except Exception as e:
            print(f"Warning: Could not verify manifest content: {e}")

    # Create the Makefile content

    makefile_content = f"""
.PHONY: build

# Allow passing tags via environment variable GO_BUILD_TAGS
TAGS_ARG := $(if $(strip $(GO_BUILD_TAGS)),-tags \"$(GO_BUILD_TAGS)\",)

build:
\tgo build $(TAGS_ARG) -ldflags \"-s -w  -extldflags '-L./lib -Wl,-rpath,./lib'\" -o {app_name} .

"""
    
    print("Creating Makefile with content:")
    print(makefile_content)

    # Write the Makefile
    with open(make_file_path, "w") as makefile:
        makefile.write(makefile_content)
    print("Makefile created successfully.", make_file_path)

if __name__ == "__main__":
    if len(sys.argv) == 3:
        app_name, manifest_file_name = sys.argv[1], sys.argv[2]
        create_makefile(app_name, ".", manifest_file_name)
    elif len(sys.argv) == 4:
        app_name, appdir, manifest_file_name = sys.argv[1], sys.argv[2], sys.argv[3]
        create_makefile(app_name, appdir, manifest_file_name)
    else:
        print("Error executing the python generate_makefile.py")
        print("Args", sys.argv, len(sys.argv))
        sys.exit(1)
    
    
