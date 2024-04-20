package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"reflect"
	"regexp"

	"github.com/Cacsjep/goxis/pkg/axmanifest"
	"github.com/erikgeiser/promptkit/textinput"
)

func Input(question, placeholder string, validator func(string) error) string {
	input := textinput.New(question)
	input.Placeholder = placeholder
	input.Validate = validator
	name, err := input.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	return name
}

func ProjectGenUi() *Project {
	modulePath := Input("Go module path", "github.com/username/repo", func(s string) error {
		if s == "" {
			return fmt.Errorf("module path cannot be empty")
		}
		validPathRegex := regexp.MustCompile(`^[\w.-]+\.[a-z]{2,}(/[\w.-]+)+$`)
		if !validPathRegex.MatchString(s) {
			return fmt.Errorf("invalid Go module path format")
		}

		return nil
	})
	appName := Input("Acap Manifest appname", "myawesomeacap", func(s string) error {
		if !regexp.MustCompile(`^\w+$`).MatchString(s) || len(s) > 26 {
			return fmt.Errorf("app name must be alphanumeric and up to 26 characters long")
		}
		return nil
	})
	friendlyName := Input("Acap Manifest friendly name", "My Awesome ACAP", func(s string) error {
		// Assuming optional but needs to be non-empty if provided
		if len(s) == 0 {
			return fmt.Errorf("friendly name must not be empty")
		}
		return nil
	})
	vendor := Input("Acap Manifest vendor name", "Company, Inc.", func(s string) error {
		if !regexp.MustCompile(`^[\w /\-\(\),\.!?\&]+$`).MatchString(s) {
			return fmt.Errorf("vendor name contains invalid characters")
		}
		return nil
	})

	return &Project{
		ModuleName:   modulePath,
		AppName:      appName,
		FriendlyName: friendlyName,
		Vendor:       vendor,
	}
}

type Project struct {
	ModuleName   string
	AppName      string
	FriendlyName string
	Vendor       string
}

func createNewProject() {

	project := ProjectGenUi()

	manifest := axmanifest.ApplicationManifestSchema{
		SchemaVersion: "1.6.0",
		ACAPPackageConf: axmanifest.ACAPPackageConf{
			Setup: axmanifest.Setup{
				AppName:            project.AppName,
				FriendlyName:       project.FriendlyName,
				Vendor:             project.Vendor,
				EmbeddedSdkVersion: "3.0",
				RunMode:            "never",
				Version:            "1.0.0",
				User:               axmanifest.User{Username: "sdk", Group: "sdk"},
			},
		},
	}

	err := os.Mkdir(project.AppName, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	samplecode := `package main

import (
	"github.com/Cacsjep/goxis/pkg/acapapp"
)

func main() {
	app := acapapp.NewAcapApplication()
	app.Syslog.Info("Hello from " + app.Manifest.ACAPPackageConf.Setup.AppName + "!")
}`
	err = os.WriteFile(path.Join(project.AppName, "main.go"), []byte(samplecode), 0755)
	if err != nil {
		fmt.Printf("Failed to write main.go: %v\n", err)
		os.Exit(1)
	}

	gitignore := `
*.eap`
	err = os.WriteFile(path.Join(project.AppName, ".gitignore"), []byte(gitignore), 0755)
	if err != nil {
		fmt.Printf("Failed to write .gitignore: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(path.Join(project.AppName, "LICENSE"), []byte("FILLME"), 0755)
	if err != nil {
		fmt.Printf("Failed to write LICENSE file: %v\n", err)
		os.Exit(1)
	}

	manContent, err := customMarshal(manifest)
	if err != nil {
		fmt.Printf("Failed to marshal manifest: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(path.Join(project.AppName, "manifest.json"), manContent, 0755)
	if err != nil {
		fmt.Printf("Failed to write manifest.json: %v\n", err)
		os.Exit(1)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	err = os.Chdir(project.AppName)
	if err != nil {
		fmt.Printf("Error changing to project directory: %v\n", err)
		os.Exit(1)
	}

	err = exec.Command("go", "mod", "init", project.ModuleName).Run()
	if err != nil {
		fmt.Printf("Failed to initialize Go module: %v\n", err)
		os.Exit(1)
	}

	err = exec.Command("go", "get", "github.com/Cacsjep/goxis").Run()
	if err != nil {
		fmt.Printf("Failed to get dependencies: %v\n", err)
		os.Exit(1)
	}

	err = os.Chdir(originalDir)
	if err != nil {
		fmt.Printf("Failed to return to original directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Project created successfully")
	fmt.Println("To build the project, run 'goxisbuilder -appdir " + project.AppName + "', or 'goxisbuilder' if you are in the project directory.")
	os.Exit(0)
}

// Custom marshaling that excludes empty fields deeply.
func customMarshal(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Create a new map to hold non-empty values.
	result := make(map[string]interface{})
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rv.Type().Field(i)
		fieldName := fieldType.Name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag != "" {
			fieldName = jsonTag
		}

		// Check for zero values, including nested structures.
		if field.Kind() == reflect.Struct {
			innerJSON, err := customMarshal(field.Interface())
			if err != nil {
				return nil, err
			}
			if len(innerJSON) > 2 { // Non-empty JSON object.
				result[fieldName] = json.RawMessage(innerJSON)
			}
		} else if !field.IsZero() {
			result[fieldName] = field.Interface()
		}
	}

	return json.MarshalIndent(result, "", "    ")
}
