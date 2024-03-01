package web

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const BaseBinariesUrl = "https://storage.googleapis.com/flutter_infra_release/releases"

func DownloadFlutterBinary(target_file_dir string, flutter_version string, channel string, os_platform string) {

	out, err := os.Create(target_file_dir + "\\v" + flutter_version + ".zip")
	if err != nil {
		log.Fatalln("Error while creating"+target_file_dir+"\\v"+flutter_version+".zip", err)
	}
	defer out.Close()

	full_download_url := fmt.Sprintf("%s/%s/%s/%s", BaseBinariesUrl, channel, os_platform, "flutter_"+os_platform+"_"+flutter_version+"-"+channel+".zip")

	resp, err := http.Get(full_download_url)

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

		err := os.RemoveAll(target_file_dir)
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
