package main

import (
	"bytes"
	"errors"
	"fmt"
	"fvm/file"
	"fvm/flutter"
	"fvm/web"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/olekukonko/tablewriter"
)

const FvmVersion = "1.0.0"

type Environment struct {
	settings        string
	root            string
	symlink         string
	flutter_mirror  string
	proxy           string
	originalpath    string
	originalversion string
	verifyssl       bool
}

var home = filepath.Clean(os.Getenv("FVM_HOME") + "\\settings.txt")
var symlink = filepath.Clean(os.Getenv("FVM_SYMLINK"))

var env = &Environment{
	settings:        home,
	root:            "",
	symlink:         symlink,
	flutter_mirror:  "",
	originalpath:    "",
	originalversion: "",
	verifyssl:       true,
}

func main() {
	args := os.Args

	detail := ""

	if !isTerminal() {
		os.Exit(0)
	}

	if len(args) > 2 {
		detail = args[2]
	}

	if len(args) < 2 {
		help()
		return
	}

	if args[1] != "version" && args[1] != "--version" && args[1] != "v" && args[1] != "-v" && args[1] != "--v" {
		setup()
	}

	switch args[1] {
	case "install":
		install(detail)
	case "use":
		use(detail)
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

func setup() {
	lines, err := file.ReadLines(env.settings)
	if err != nil {
		fmt.Println("\nERROR", err)
		os.Exit(1)
	}

	// Process each line and extract the value
	m := make(map[string]string)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = os.ExpandEnv(line)
		res := strings.Split(line, ":")
		if len(res) < 2 {
			continue
		}
		m[res[0]] = strings.TrimSpace(strings.Join(res[1:], ":")) //In case that the value is filepath with e.g C:\Users, join all path elements to the complete string
	}

	if val, ok := m["root"]; ok {
		env.root = filepath.Clean(val)
	}
	if val, ok := m["originalpath"]; ok {
		env.originalpath = filepath.Clean(val)
	}
	if val, ok := m["originalversion"]; ok {
		env.originalversion = val
	}
	if val, ok := m["flutter_mirror"]; ok {
		env.flutter_mirror = val
	}

	web.SetMirrors(env.flutter_mirror)

	// Make sure the directories exist
	_, e := os.Stat(env.root)
	if e != nil {
		fmt.Println(env.root + " could not be found or does not exist. Exiting.")
		return
	}
}

func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func splitVersion(version string) map[string]int {
	parts := strings.Split(version, ".")
	var result = make([]int, 3)

	for i, item := range parts {
		v, _ := strconv.Atoi(item)
		result[i] = v
	}

	return map[string]int{
		"major": result[0],
		"minor": result[1],
		"patch": result[2],
	}
}

func findLatestSubVersion(version string, localOnly ...bool) string {
	if len(localOnly) > 0 && localOnly[0] {
		installed := flutter.GetInstalled(env.root)
		result := ""
		for _, v := range installed {
			if !strings.HasPrefix(v, "v"+version) {
				continue
			}

			if result == "" {
				result = v
				continue
			}

			current, _ := semver.New(result)
			next, _ := semver.New(v)
			if current.LT(*next) {
				result = v
			}

		}

		if len(strings.TrimSpace(result)) > 0 {
			return result
		}
	}

	_, _, stables, betas := flutter.GetReleases()
	all := append(stables, betas...)
	if len(strings.Split(version, ".")) == 2 {
		requestedVersion := splitVersion(version + ".0")
		for _, v := range all {
			available := splitVersion(v)
			if requestedVersion["major"] == available["major"] {
				if requestedVersion["minor"] == available["minor"] {
					if available["patch"] > requestedVersion["patch"] {
						requestedVersion["patch"] = available["patch"]
					}
				}
				if requestedVersion["minor"] > available["minor"] {
					break
				}
			}

			if requestedVersion["major"] > available["major"] {
				break
			}
		}
		return fmt.Sprintf("%v.%v.%v", requestedVersion["major"], requestedVersion["minor"], requestedVersion["patch"])
	}

	semverVersions := []semver.Version{}
	majorVersion := strings.Split(version, ".")[0]
	for _, newVersion := range all {
		if strings.HasPrefix(newVersion, majorVersion) {
			semVersion, _ := semver.Make(newVersion)
			semverVersions = append(semverVersions, semVersion)
		}
	}

	if len(semverVersions) == 0 {
		return ""
	}

	semver.Sort(semverVersions)

	return semverVersions[len(semverVersions)-1].String()
}

func getVersion(version string, localInstallsOnly ...bool) (string, error) {
	requestedVersion := version
	lts, current, _, _ := flutter.GetReleases()
	if version == "" {
		return "", errors.New("a version argument is required but missing")
	}

	if version == "latest" {
		version = current[0]
		fmt.Println(version)
	}

	if version == "lts" {
		version = lts[0]
		fmt.Println(version)
	}

	if version == "newest" {
		installed := flutter.GetInstalled(env.root)
		if len(installed) == 0 {
			return version, errors.New("no versions of node.js found. Try installing the latest by typing nvm install latest")
		}

		version = installed[0]
	}

	v, err := semver.Make(version)

	if err == nil {
		err = v.Validate()
	}

	if err == nil {
		// if the user specifies only the major/minor version, identify the latest
		// version applicable to what was provided.
		sv := strings.Split(version, ".")
		if len(sv) < 3 {
			version = findLatestSubVersion(version)
		}
	} else if strings.Contains(err.Error(), "No Major.Minor.Patch") {
		latestLocalInstall := false
		if len(localInstallsOnly) > 0 {
			latestLocalInstall = localInstallsOnly[0]
		}
		version = findLatestSubVersion(version, latestLocalInstall)
		if len(version) == 0 {
			err = errors.New("Unrecognized version: \"" + requestedVersion + "\"")
		}
	}

	return version, err
}

func accessDenied(err error) bool {
	fmt.Println(err)

	return strings.Contains(strings.ToLower(err.Error()), "access is denied")
}

func run(cmd string, dir *string, arg ...string) (bool, error) {
	c := exec.Command(cmd, arg...)
	if dir != nil {
		c.Dir = *dir
	}
	var stderr bytes.Buffer
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		return false, errors.New(fmt.Sprint(err) + ": " + stderr.String())
	}

	return true, nil
}

func elevatedRun(command string, args ...string) (bool, error) {
	ok, err := run("cmd", nil, append([]string{"/C", command}, args...)...)

	if err != nil {
		ok, err = run("elevate.cmd", &env.root, append([]string{"cmd", "/C", command}, args...)...)
	}

	return ok, err
}

func use(version string) {
	version, err := getVersion(version)

	if err != nil {
		if !strings.Contains(err.Error(), "No Major.Minor.Patch") {
			fmt.Println(err.Error())
			return
		}
	}

	if !flutter.IsVersionInstalled(env.root, version) {
		fmt.Println("flutter v" + version + " is not installed.")
		return
	}

	flutterRootDir := filepath.Join(filepath.Clean(env.symlink), "..")

	sym, _ := os.Lstat(flutterRootDir)
	if sym != nil {
		_, err := elevatedRun("rmdir", flutterRootDir, "/S", "/Q")

		if err != nil {
			if accessDenied(err) {
				return
			}
		}
	}

	var ok bool

	ok, err = elevatedRun("mklink", "/D", flutterRootDir, filepath.Join(env.root, "v"+version))
	if err != nil {
		if strings.Contains(err.Error(), "not have sufficient privilege") || strings.Contains(strings.ToLower(err.Error()), "access is denied") {
			ok, err = elevatedRun("mklink", "/D", flutterRootDir, filepath.Join(env.root, "v"+version))
			if err != nil {
				fmt.Println(fmt.Sprint(err))
			}
		} else if strings.Contains(err.Error(), "file already exists") {
			ok, err = elevatedRun("rmdir", flutterRootDir, "/S", "/Q")
			reloadable := true
			if err != nil {
				fmt.Println(fmt.Sprint(err))
			} else if reloadable {
				use(version)
				return
			}
		} else {
			fmt.Print(fmt.Sprint(err))
		}
	}

	if !ok {
		return
	}

	fmt.Println("Now using flutter v" + version)

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
		fmt.Println("")
		currentVersion := flutter.GetCurrentVersion()
		installedVersions := flutter.GetInstalled(env.root)

		for _, version := range installedVersions {
			isFlutter := regexp.MustCompile("v").MatchString(version)

			if !isFlutter {
				continue
			}

			version = strings.Replace(version, "v", "", 1)
			if currentVersion == version {
				fmt.Println(version + " *")
				continue
			}
			fmt.Println(version + " ")
		}

		if len(installedVersions) == 0 {
			fmt.Println("No installations recognized.")
		}
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
