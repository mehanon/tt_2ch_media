package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cavaliergopher/grab/v3"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("%v", err)
			println("\nPress any button to exit...")
			_, _ = bufio.NewReader(os.Stdin).ReadByte()
		}
	}()

	if len(os.Args) > 1 { // with console args
		for _, arg := range os.Args[1:] {
			if arg == "-h" || arg == "--help" {
				println("note: windows firewall could block the script from accessing the internet")
				println("source code: https://github.com/mehanon/tt_2ch_media")
				println("usage example:")
				println("1. In terminal -- ./tt_2ch_media https://www.tiktok.com/@shrimpydimpy/video/7133412834960018730")
				println("2. Inline mode -- ./tt_2ch_media")
				continue
			}

			filename, err := DownloadTiktokTikwm(arg)
			if err != nil {
				log.Fatalf("While getting %s, an error occured:\n%s", arg, err.Error())
			}
			fmt.Printf("%s\n", *filename)
		}
	} else { // like classic application
		println("just enter down TikTok links to download (or an empty line for exit)")
		println("  note: windows firewall could block the script from accessing the internet")
		println("  source code: https://github.com/mehanon/tt_2ch_media")

		reader := bufio.NewReader(os.Stdin)
		for {
			links, err := reader.ReadString('\n')
			links = links[:len(links)-1]
			if err != nil {
				log.Fatalf("While reading input, an error occured:\n%s", err.Error())
			}
			for _, link := range strings.Split(links, " ") {
				link = strings.Trim(link, " \n\t\"'")
				if len(link) == 0 {
					println("see you next time")
					time.Sleep(time.Second * 5)
					os.Exit(0)
				}
				fmt.Printf("  wokring on %s...\n", link)
				filename, err := DownloadTiktokTikwm(link)
				if err != nil {
					log.Fatalf("While getting %s, an error occured:\n%s", link, err.Error())
				}
				fmt.Printf("done -> %s\n", *filename)
			}
		}
	}
}

// TiktokInfo there are more fields, tho I omitted unnecessary ones
type TiktokInfo struct {
	Id         string `json:"id"`
	Play       string `json:"play,omitempty"`
	Hdplay     string `json:"hdplay,omitempty"`
	CreateTime int64  `json:"create_time"`
	Author     struct {
		UniqueId string `json:"unique_id"`
	} `json:"author"`
}

type TikwmResponse struct {
	Code          int        `json:"code"`
	Msg           string     `json:"msg"`
	ProcessedTime float64    `json:"processed_time"`
	Data          TiktokInfo `json:"data,omitempty"`
}

func TikwnGetInfo(link string) (*TiktokInfo, error) {
	payload := url.Values{"url": {link}, "hd": {"1"}}
	r, err := http.PostForm("https://www.tikwm.com/api/", payload)
	if err != nil {
		return nil, err
	}
	buffer, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var resp TikwmResponse
	err = json.Unmarshal(buffer, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Msg)
	}

	return &resp.Data, nil
}

func DownloadTiktokTikwm(link string) (*string, error) {
	info, err := TikwnGetInfo(link)
	if err != nil {
		return nil, err
	}

	var downloadUrl string
	if info.Hdplay != "" {
		downloadUrl = info.Hdplay
	} else if info.Play != "" {
		println("warning: tikwm couldn't find HD version, downloading how it is...")
		downloadUrl = info.Play
	} else {
		return nil, errors.New("no download links found :c")
	}

	localFilename := fmt.Sprintf(
		"%s_%s_%s.mp4",
		info.Author.UniqueId,
		time.Unix(info.CreateTime, 0).Format("2006-01-02"),
		info.Id,
	)

	err = Wget(downloadUrl, localFilename)
	if err != nil {
		return nil, err
	}

	return &localFilename, nil
}

func Wget(url string, filename string) error {
	_, err := grab.Get(filename, url)
	return err
}
