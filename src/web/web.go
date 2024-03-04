package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type Releases struct {
	Hash           string    `json:"hash"`
	Channel        string    `json:"channel"`
	Version        string    `json:"version"`
	DartSdkVersion string    `json:"dart_sdk_version,omitempty"`
	DartSdkArch    string    `json:"dart_sdk_arch,omitempty"`
	ReleaseDate    time.Time `json:"release_date"`
	Archive        string    `json:"archive"`
	Sha256         string    `json:"sha256"`
}

type FlutterReleases struct {
	BaseURL        string         `json:"base_url"`
	CurrentRelease CurrentRelease `json:"current_release"`
	Releases       []Release      `json:"releases"`
}

type CurrentRelease struct {
	Beta   string `json:"beta"`
	Dev    string `json:"dev"`
	Stable string `json:"stable"`
}

type Release struct {
	Archive        string    `json:"archive"`
	Channel        string    `json:"channel"`
	DartSdkArch    string    `json:"dart_sdk_arch,omitempty"`
	DartSdkVersion string    `json:"dart_sdk_version,omitempty"`
	Hash           string    `json:"hash"`
	ReleaseDate    time.Time `json:"release_date"`
	Sha256         string    `json:"sha256"`
	Version        string    `json:"version"`
}

const BaseBinariesUrl = "https://storage.googleapis.com/flutter_infra_release/releases"

const FlutterReleasesUrl = "https://storage.googleapis.com/flutter_infra_release/releases/releases_windows.json"

func GetAllReleases() []Release {
	response, err := http.Get(FlutterReleasesUrl)

	if err != nil {
		log.Fatalln(err)
	}

	defer response.Body.Close()

	var flutterReleases FlutterReleases

	if err := json.NewDecoder(response.Body).Decode(&flutterReleases); err != nil {
		log.Fatalln(err)
	}

	return flutterReleases.Releases
}

func DownloadFlutterBinary(targetFileDir string, flutterVersion string, channel string, osPlatform string) {

	targetZipFile := filepath.Join(targetFileDir, "v"+flutterVersion+".zip")

	out, err := os.Create(targetZipFile)
	if err != nil {
		log.Fatalln("Error while creating: "+targetZipFile+" ", err)
	}
	defer out.Close()

	fullDownloadUrl := fmt.Sprintf("%s/%s/%s/%s", BaseBinariesUrl, channel, osPlatform, "flutter_"+osPlatform+"_"+flutterVersion+"-"+channel+".zip")

	resp, err := http.Get(fullDownloadUrl)

	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("Download interrupted. Rolling back...")
		out.Close()
		resp.Body.Close()

		err := os.RemoveAll(targetFileDir)
		if err != nil {
			fmt.Println("Error while rolling back", err)
		}
		os.Exit(1)
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

}
