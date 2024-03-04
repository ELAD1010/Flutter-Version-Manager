package main

import (
	"fmt"
	"fvm/file"
	"fvm/flutter"
	"fvm/web"
	"os"
	"path/filepath"
	"runtime"

	"github.com/olekukonko/tablewriter"
)

const FvmVersion = "1.0.0"

func main() {
	args := os.Args

	detail := ""

	if len(args) > 2 {
		detail = args[2]
	}

	if len(args) < 2 {
		help()
		return
	}

	switch args[1] {
	case "install":
		install(detail)
	case "ls":
		fallthrough
	case "list":
		list(detail)
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

func list(filter string) {

	if filter == "" {
		filter = "installed"
	}

	if filter != "installed" && filter != "available" {
		fmt.Println("\nInvalid list option.\n\nPlease use on of the following\n  - nvm list\n  - nvm list installed\n  - nvm list available")
		help()
		return
	}

	if filter == "installed" {

	}

	if filter == "available" {
		showReleases()
	}

}

func showReleases() {

	lts, current, stables, betas := flutter.GetReleases()

	releases := 25

	data := make([][]string, releases, releases+5)

	for i := 0; i < releases; i++ {
		release := make([]string, 4, 6)

		release[0] = ""
		release[1] = ""
		release[2] = ""
		release[3] = ""

		if len(current) > i {
			if len(current[i]) > 0 {
				release[0] = current[i]
			}
		}

		if len(lts) > i {
			if len(lts[i]) > 0 {
				release[1] = lts[i]
			}
		}

		if len(stables) > i {
			if len(stables[i]) > 0 {
				release[2] = stables[i]
			}
		}

		if len(betas) > i {
			if len(betas[i]) > 0 {
				release[3] = betas[i]
			}
		}

		data[i] = release
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"   Current  ", "    LTS     ", " Stable ", "Beta"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()

	fmt.Println("\nThis is a partial list. For a complete list, visit https://docs.flutter.dev/release/archive")

}

func install(flutterVersion string) {
	targetFileDir := filepath.Join(os.Getenv("FVM_HOME"), "v"+flutterVersion)
	targetZipFile := filepath.Join(targetFileDir, "v"+flutterVersion+".zip")
	os.Mkdir(targetFileDir, os.ModeDir)

	channelType := flutter.GetChannelType(flutterVersion)

	osPlatform := runtime.GOOS

	if osPlatform == "darwin" {
		osPlatform = "macos"
	}

	web.DownloadFlutterBinary(targetFileDir, flutterVersion, channelType, osPlatform)

	file.Unzip(targetZipFile, targetFileDir)
	os.Remove(targetZipFile)

	fmt.Println("\n\nInstallation complete. If you want to use this version, type\n\nfvm use " + flutterVersion)
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
