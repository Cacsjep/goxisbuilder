package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/icholy/digest"
)

const (
	Blue  = "\033[34m"
	Reset = "\033[0m"
	Green = "\033[32m"
)

func boolToStr(b bool) string {
	if b {
		return "YES"
	}
	return "NO"
}

func ptr(s string) *string {
	return &s
}

func getLog(url string, pwd string) {
	client := &http.Client{
		Transport: &digest.Transport{
			Username: "root",
			Password: pwd,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("FETCH LOG ERROR: %s", err)
		return
	}

	if resp.StatusCode == 401 {
		log.Println("FETCH LOG ERROR: Unauthorized, either the password is incorrect or the camera does not support digest auth -> check root.Network.HTTP.AuthenticationPolicy via https://<ip>/axis-cgi/param.cgi?action=list, should be set to digest")
		return
	}

	if resp.StatusCode != 200 {
		log.Printf("FETCH LOG ERROR: %s", resp.Status)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("FETCH LOG ERROR: %s", err)
		return
	}

	displayLastLines(string(body), 70)
}

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func displayLastLines(logContent string, nline int) {
	fmt.Println("Update..")
	time.Sleep(time.Millisecond * 500)
	clearScreen()
	lines := strings.Split(logContent, "\n")
	startLine := 0
	if len(lines) > nline {
		startLine = len(lines) - nline
	}
	for _, line := range lines[startLine:] {
		fmt.Println(line)
	}
}

// handleError logs an error message and exits the program with a status code.
func handleError(message string, err error) {
	log.Printf("Error: %s: %v\n", message, err)
	os.Exit(1) // Exit with a status code indicating failure.
}

// configureArchitecture sets up the build configuration based on the architecture.
func configureArchitecture(arch string, buildConfig *BuildConfiguration) {
	switch arch {
	case "aarch64":
		buildConfig.ImageName = "acap:aarch64"
		buildConfig.GoArch = "arm64"
		buildConfig.CrossPrefix = "aarch64-linux-gnu-"
	case "armv7hf":
		buildConfig.ImageName = "acap:arm"
		buildConfig.GoArch = "arm"
		buildConfig.GoArm = "7"
		buildConfig.CrossPrefix = "arm-linux-gnueabihf-"
	default:
		handleError("Architecture invalid", fmt.Errorf("should be either aarch64 or armv7hf, got %s", arch))
	}
}

// configureSdk sets up the build configuration based on the lowest Sdk flag.
func configureSdk(lowestSdkVersion bool, buildConfig *BuildConfiguration) {
	if lowestSdkVersion {
		buildConfig.Sdk = "acap-sdk"
		buildConfig.UbunutVersion = "20.04"
		if buildConfig.SdkVersion != "" {
			buildConfig.Version = buildConfig.SdkVersion
		} else {
			buildConfig.Version = "3.5"
		}
	} else {
		buildConfig.Sdk = "acap-native-sdk"
		if buildConfig.UbunutVersion == "" {
			buildConfig.UbunutVersion = "24.04"
		}
		if buildConfig.SdkVersion != "" {
			buildConfig.Version = buildConfig.SdkVersion
		} else {
			buildConfig.Version = "12.7.0"
		}
	}
	fmt.Println("Using SDK:", buildConfig.Sdk)
	fmt.Println("Using SDK version:", buildConfig.Version)
	fmt.Println("Using Ubuntu version:", buildConfig.UbunutVersion)
}

func watchPackageLog(buildConfig *BuildConfiguration) {
	// Setup a channel to listen for interrupt signal (Ctrl+C)
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	url := fmt.Sprintf("https://%s/axis-cgi/admin/systemlog.cgi?appname=%s", buildConfig.Ip, buildConfig.Manifest.ACAPPackageConf.Setup.AppName)
Loop:
	for {
		select {
		case <-ticker.C:
			getLog(url, buildConfig.Pwd)
		case <-sigChan:
			fmt.Println("Interrupt received, stopping...")
			break Loop
		}
	}
}

func listEapDirectory() {
	entries, err := os.ReadDir("./build")
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		fmt.Println("EAP:", e.Name())
	}
}

func printCompatibility(buildConfig *BuildConfiguration) {
	fmt.Println("\n\nAcap Compatibility:")
	// Maps for SDK to Firmware compatibility
	sdkToFirmware := map[string]string{
		"3.0": "9.70 and later",
		"3.1": "9.80 (LTS) and later",
		"3.2": "10.2 and later",
		"3.3": "10.5 and later",
		"3.4": "10.6 and later",
		"3.5": "10.9 and later",
	}

	// Maps for Native SDK version to Firmware compatibility
	nativeSdkToFirmware := map[string]string{
		"1.0":    "10.7 and later until LTS",
		"1.1":    "10.9 and later until LTS",
		"1.2":    "10.10 and later until LTS",
		"1.3":    "10.12 (LTS)",
		"1.4":    "11.0 and later until LTS",
		"1.5":    "11.1 and later until LTS",
		"1.6":    "11.2 and later until LTS",
		"1.7":    "11.3 and later until LTS",
		"1.8":    "11.4 and later until LTS",
		"1.9":    "11.5 and later until LTS",
		"1.10":   "11.6 and later until LTS",
		"1.11":   "11.7 and later until LTS",
		"1.12":   "11.8 and later until LTS",
		"1.13":   "11.9 and later until LTS",
		"1.14":   "11.10 and later until LTS",
		"1.15":   "11.11 (LTS)",
		"12.0.0": "12.0 and later until LTS",
		"12.1.0": "12.1 and later until LTS",
		"12.2.0": "12.2 and later until LTS",
		"12.3.0": "12.2 and later until LTS",
		"12.4.0": "12.2 and later until LTS",
		"12.5.0": "12.2 and later until LTS",
		"12.6.0": "12.2 and later until LTS",
		"12.7.0": "12.7 and later until LTS",
	}

	// Check if it's using the native SDK or standard SDK
	if buildConfig.Sdk == "acap-native-sdk" {
		if firmware, ok := nativeSdkToFirmware[buildConfig.Version]; ok {
			fmt.Printf("     ACAP Native SDK %s%s%s, compatible with AXIS OS version: %s%s%s\n", Blue, buildConfig.Version, Reset, Green, firmware, Reset)
		} else {
			log.Printf("     Unknown ACAP Native SDK version: %s\n", buildConfig.Version)
		}
	} else if buildConfig.Sdk == "acap-sdk" {
		if firmware, ok := sdkToFirmware[buildConfig.Version]; ok {
			fmt.Printf("     ACAP3 SDK %s%s%s, compatible with firmware version: %s%s%s\n", Blue, buildConfig.Version, Reset, Green, firmware, Reset)
		} else {
			log.Printf("     Unknown ACAP3 SDK version: %s\n", buildConfig.Version)
		}
	} else {
		log.Printf("     Unknown SDK configuration: %s\n", buildConfig.Sdk)
	}

	schemaToFirmware := map[string]string{
		"1.0":   "10.7",
		"1.1":   "10.7",
		"1.2":   "10.7",
		"1.3":   "10.9",
		"1.3.1": "11.0",
		"1.4.0": "11.7",
		"1.5.0": "11.8",
		"1.6.0": "11.9",
		"1.7.0": "11.10",
		"1.7.1": "12.0",
		"1.7.2": "12.1",
		"1.7.3": "12.2",
		"1.7.4": "12.4",
		"1.8.0": "12.6",
	}

	if firmware, ok := schemaToFirmware[buildConfig.Manifest.SchemaVersion]; ok {
		fmt.Printf("     Schema %s%s%s is compatible with firmware version: %s%s%s\n", Blue, buildConfig.Manifest.SchemaVersion, Reset, Green, firmware, Reset)
	} else {
		log.Printf("     Unknown Schema version: %s\n", buildConfig.Manifest.SchemaVersion)
	}

	archToChips := map[string][]string{
		"armv7hf": {"ARTPEC-6", "ARTPEC-7", "i.MX 6SoloX", "i.MX 6ULL"},
		"aarch64": {"ARTPEC-8", "CV25", "S5", "S5L"},
	}

	if chips, ok := archToChips[buildConfig.Arch]; ok {
		chipsStr := strings.Join(chips, ", ")

		fmt.Printf("     Supported architecture: %s%s%s with chips: %s%s%s\n",
			Blue, buildConfig.Arch, Reset,
			Green, chipsStr, Reset)
	} else {
		fmt.Println("     Unsupported architecture.")
	}
}

// normalizeGoBuildTags converts a user-provided tags string into a
// comma-separated list per modern Go expectations. It accepts input in
// either space- or comma-separated form, trims whitespace, removes empty
// entries, and de-duplicates while preserving order.
func normalizeGoBuildTags(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Treat commas as separators, then split on any whitespace
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return ""
	}
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return strings.Join(out, ",")
}
