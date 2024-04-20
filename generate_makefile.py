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
            os.remove(default_manifest)
        os.rename(custom_manifest, default_manifest)
        print(f"Manifest renamed from {manifest_file_name} to {default_manifest}")

    # Create the Makefile content
    makefile_content = f"""
.PHONY: build
build:
\tgo build -ldflags "-s -w  -extldflags '-L./lib -Wl,-rpath,./lib'" -o {app_name} .
"""

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
    
    