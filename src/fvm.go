package main

import (
	"fmt"
	"fvm/file"
	"fvm/web"
	"os"
	"path/filepath"
)

const FvmVersion = "1.0.0"

func main() {
	args := os.Args

	if len(args) < 2 {
		help()
		return
	}

	switch args[1] {
	case "install":
		install(args[2])
	case "v":
		fmt.Println(FvmVersion)
	case "--version":
		fallthrough
	case "--v":
		fallthrough
	case "-v":
		fallthrough
	case "version":
		fmt.Println(FvmVersion)
	default:
		help()
	}
}

func install(flutter_version string) {
	target_file_dir := filepath.Join(os.Getenv("FVM_HOME"), "v"+flutter_version)
	target_zip_file := filepath.Join(target_file_dir, "v"+flutter_version+".zip")
	os.Mkdir(target_file_dir, os.ModeDir)
	web.DownloadFlutterBinary(target_file_dir, flutter_version, "stable", "windows")

	file.Unzip(target_zip_file, target_file_dir)
	os.Remove(target_zip_file)

	fmt.Println("\n\nInstallation complete. If you want to use this version, type\n\nfvm use " + flutter_version)
}

func help() {
	fmt.Println("\nRunning version " + FvmVersion + ".")
	fmt.Println("\nUsage:")
	fmt.Println(" ")
	fmt.Println("  fvm arch                     : Show if flutter is running in 32 or 64 bit mode.")
	fmt.Println("  fvm current                  : Display active version.")
	fmt.Println("  fvm install <version> [arch] : The version can be a specific version, \"latest\" for the latest current version, or \"lts\" for the")
	fmt.Println("                                 most recent LTS version. Optionally specify whether to install the 32 or 64 bit version (defaults")
	fmt.Println("                                 to system arch). Set [arch] to \"all\" to install 32 AND 64 bit versions.")
	fmt.Println("                                 Add --insecure to the end of this command to bypass SSL validation of the remote download server.")
	fmt.Println("  fvm list [available]         : List the node.js installations. Type \"available\" at the end to see what can be installed. Aliased as ls.")
	fmt.Println("  fvm on                       : Enable flutter version management.")
	fmt.Println("  fvm off                      : Disable flutter version management.")
	fmt.Println("  fvm proxy [url]              : Set a proxy to use for downloads. Leave [url] blank to see the current proxy.")
	fmt.Println("                                 Set [url] to \"none\" to remove the proxy.")
	fmt.Println("  fvm flutter_mirror [url]        : Set the node mirror. Defaults to https://storage.googleapis.com/flutter_infra_release/releases/. Leave [url] blank to use default url.")
	fmt.Println("  fvm uninstall <version>      : The version must be a specific version.")
	//  fmt.Println("  nvm update                   : Automatically update nvm to the latest version.")
	fmt.Println("  fvm use [version] [arch]     : Switch to use the specified version. Optionally use \"latest\", \"lts\", or \"newest\".")
	fmt.Println("                                 \"newest\" is the latest installed version. Optionally specify 32/64bit architecture.")
	fmt.Println("                                 nvm use <arch> will continue using the selected version, but switch to 32/64 bit mode.")
	fmt.Println("  fvm root [path]              : Set the directory where nvm should store different versions of node.js.")
	fmt.Println("                                 If <path> is not set, the current root will be displayed.")
	fmt.Println("  fvm [--]version              : Displays the current running version of nvm for Windows. Aliased as v.")
	fmt.Println(" ")
}
