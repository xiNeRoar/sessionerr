package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	qbittorrent "github.com/autobrr/go-qbittorrent"
)

func main() {
	var password, host, username, path, crossSeedHost string
	flag.StringVar(&username, "U", os.Getenv("USERNAME"), "username")
	flag.StringVar(&host, "H", os.Getenv("HOST"), "host")
	flag.StringVar(&password, "P", os.Getenv("PASSWORD"), "password")
	flag.StringVar(&path, "S", os.Getenv("SESSIONDIR"), "BT_backup")
	flag.StringVar(&crossSeedHost, "C", os.Getenv("CROSSSEED"), "CROSSSEED URL")

	flag.Parse()

	base := filepath.Join("./", path, "/BT_backup") + "/"

	c := qbittorrent.NewClient(qbittorrent.Config{
		Host:     host,
		Username: username,
		Password: password,
	})

	if err := c.Login(); err != nil {
		fmt.Printf("Unable to login: %q\n", err)
		os.Exit(1)
	}

	torrents, err := c.GetTorrents(qbittorrent.TorrentFilterOptions{})
	if err != nil {
		fmt.Printf("Unable to get Torrent List: %q\n", err)
		os.Exit(2)
	}

	files, err := filepath.Glob(base + "*")
	if err != nil {
		fmt.Printf("Unable to get Backup dir: %q\n", err)
		os.Exit(3)
	}

	for _, k := range files {
		safe := false
		for _, t := range torrents {
			if strings.Contains(k, t.Hash) {
				safe = true
				break
			}
		}

		if !safe {
			fmt.Printf("Cleaning: %q\n", k)
			os.Remove(k)
		}
	}

	var wg sync.WaitGroup
	for _, k := range torrents {
		exists := false
		for _, v := range files {
			if strings.Contains(v, k.Hash) {
				exists = true
				break
			}
		}

		if exists {
			continue
		}

		data, err := c.ExportTorrent(k.Hash)
		if err != nil {
			fmt.Printf("Unable to export Hash %q: %q\n", k.Hash, err)
			continue
		}

		if err := os.WriteFile(base+k.Hash+".torrent", data, 0644); err != nil {
			fmt.Printf("Unable to write Hash %q: %q\n", k.Hash, err)
			os.Remove(base + k.Hash + ".torrent")
			continue
		}

		wg.Add(1)
		go func(hash, u string) {
			if err := NotifyCrosseed(hash, u); err != nil {
				fmt.Printf("Cross-Seed submission failed (%q) (%q): %q\n", hash, u, err)
			}
			wg.Done()
		}(k.Hash, crossSeedHost)
	}

	wg.Wait()
}

func NotifyCrosseed(hash, urlStr string) error {
	if len(urlStr) == 0 {
		return nil
	}

	var u url.Values
	u.Set("infoHash", hash)

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(u.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}
