package flutter

import (
	"fvm/file"
	"fvm/web"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/blang/semver"
)

func reverseStringArray(str []string) []string {
	for i := 0; i < len(str)/2; i++ {
		j := len(str) - i - 1
		str[i], str[j] = str[j], str[i]
	}

	return str
}

func GetCurrentVersion() string {
	cmd := exec.Command("flutter", "--version")
	str, err := cmd.Output()
	if err == nil {
		flutterVersionDescription, _, found := strings.Cut(string(str), "•")

		v := strings.TrimSpace(strings.TrimPrefix(flutterVersionDescription, "Flutter"))

		if !found {
			return "Unknown"
		}

		return v
	}
	return "Unknown"
}

func IsVersionInstalled(root string, version string) bool {
	isFlutterVersionExists := file.Exists(root + "\\v" + version + "\\bin\\flutter.bat")
	return isFlutterVersionExists
}

func GetInstalled(root string) []string {
	list := make([]semver.Version, 0)
	files, _ := os.ReadDir(root)

	for i := len(files) - 1; i >= 0; i-- {
		if !files[i].IsDir() && !(files[i].Type()&os.ModeSymlink == os.ModeSymlink) {
			continue
		}

		pattern, _ := regexp.Compile("v")

		isFlutter := pattern.MatchString(files[i].Name())

		if isFlutter {
			currentVersionString := strings.Replace(files[i].Name(), "v", "", 1)
			currentVersion, _ := semver.Make(currentVersionString)

			list = append(list, currentVersion)
		}

	}

	semver.Sort(list)

	installedList := make([]string, 0)

	for _, version := range list {
		installedList = append(installedList, "v"+version.String())
	}

	installedList = reverseStringArray(installedList)

	return installedList
}

func GetChannelType(version string) string {
	if strings.Contains(version, "pre") {
		return "beta"
	}
	return "stable"
}

func GetReleases() ([]string, []string, []string, []string) {
	flutterReleases := web.GetAllReleases()

	lts, current, stables, betas := []string{}, []string{}, []string{}, []string{}

	for i := 0; i < len(flutterReleases); i++ {

		if strings.Contains(flutterReleases[i].Version, "hotfix") {
			continue
		}
		if flutterReleases[i].Channel == "stable" {
			version := strings.TrimLeft(flutterReleases[i].Version, "v")
			stables = append(stables, version)
		}
		if flutterReleases[i].Channel == "beta" {
			version := strings.TrimLeft(flutterReleases[i].Version, "v")
			betas = append(betas, version)
		}
	}

	current = append(current, betas[0])
	lts = append(lts, stables[0])

	return lts, current, stables, betas
}
