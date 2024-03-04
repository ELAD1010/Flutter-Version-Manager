package flutter

import (
	"fvm/web"
	"strings"
)

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
