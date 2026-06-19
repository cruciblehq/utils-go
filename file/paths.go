package file

import "path/filepath"

// Standard filenames used by resource projects.
const (

	// Standard filename for Crucible manifests.
	ManifestFile = "crucible.yaml"

	// Standard filename for OCI image tarballs produced by runtime and service builds.
	ImageFile = "image.tar"

	// Filename for the compiled non-OCI affordance artifact.
	//
	// Holds the encoded affordance sections (MAC, network, file capabilities,
	// kernel, volume) that are attached to the image as an affordance artifact
	// for the runtime enforcement plugin.
	AffordanceFile = "affordance.json"

	// Standard filename for the JavaScript bundle produced by widget builds.
	WidgetMainFile = "index.js"

	// Standard filename for the deployment plan produced by blueprint builds.
	PlanFile = "plan.yaml"

	// Standard filename for the packaged resource archive.
	PackageFile = "package.tar.zst"
)

// Standard subdirectory names used by resource projects.
const (

	// Subdirectory name for build artifacts within a project directory.
	BuildDirName = "build"

	// Subdirectory name for distributable archives within a project directory.
	DistDirName = "dist"
)

// Path to the build output directory for a resource project.
//
// Contains built artifacts whose contents depend on the resource type. For
// example, widgets produce compiled JavaScript bundles while services produce
// OCI image tarballs.
func BuildDir(base string) string {
	return filepath.Join(base, BuildDirName)
}

// Path to the dist output directory for a resource project.
//
// Contains packaged archives and generated plans.
func DistDir(base string) string {
	return filepath.Join(base, DistDirName)
}

// Path to the default package archive for a resource project.
func Package(base string) string {
	return filepath.Join(DistDir(base), PackageFile)
}

// Path to the manifest file for a resource project.
func Manifest(base string) string {
	return filepath.Join(base, ManifestFile)
}

// Path to the deployment plan file for a resource project.
func Plan(base string) string {
	return filepath.Join(base, PlanFile)
}

// Path to the JavaScript bundle entry point for a resource project.
func WidgetMain(base string) string {
	return filepath.Join(base, WidgetMainFile)
}
