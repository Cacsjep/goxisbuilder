package main

import "github.com/Cacsjep/goxis/pkg/axmanifest"

// BuildConfiguration defines the configuration parameters for building
// the EAP application, including details such as architecture, manifest details,
// and flags indicating whether to install the application, start it,
// build examples, or watch logs.
type BuildConfiguration struct {
	Manifest        *axmanifest.ApplicationManifestSchema
	ManifestPath    string
	ImageName       string
	Ip              string
	Pwd             string
	Arch            string
	DoStart         bool
	DoInstall       bool
	Prune           bool
	AppDirectory    string
	Sdk             string
	UbunutVersion   string
	Version         string
	GoArch          string
	GoArm           string
	CrossPrefix     string
	LowestSdk       bool
	Watch           bool
	Dockerfile      string
	FilesToAdd      string
	SdkVersion      string
	ExtraLibsScript string
}
