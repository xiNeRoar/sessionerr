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
	"time"

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

	// Decoding the URL-encoded password
	decodedPassword, err := url.QueryUnescape(password)
	if err != nil {
		fmt.Printf("Error decoding password: %v\n", err)
		os.Exit(1)
	}

	c := qbittorrent.NewClient(qbittorrent.Config{
		Host:     host,
		Username: username,
		Password: decodedPassword,
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

	base := filepath.Join("./", path, "/BT_backup") + "/"
	files, err := filepath.Glob(base + "*")
	if err != nil {
		fmt.Printf("Unable to get Backup dir: %q\n", err)
		os.Exit(3)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go CleanSessionDir(base, torrents, files, &wg)
	go SubmitSessionTorrents(base, crossSeedHost, torrents, files, c, &wg)

	wg.Wait()
}

func SubmitSessionTorrents(base, crossSeedHost string, torrents []qbittorrent.Torrent, files []string, c *qbittorrent.Client, wg *sync.WaitGroup) {
	for _, k := range torrents {
		if k.Progress < 100.0 {
			continue
		}

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

		if len(crossSeedHost) != 0 && !strings.Contains(k.Tags, "cross-seed") {
			/* Works around a very silly Cross-Seed bug where it throws an internal error if it already knows about the infohash. */
			wg.Add(1)
			go func(hash, u string) {
				if err := NotifyCrosseed(hash, u); err != nil {
					fmt.Printf("Cross-Seed submission failed (%q) (%q): %q\n", hash, u, err)
				}
				wg.Done()
			}(k.Hash, crossSeedHost)
		}
	}

	wg.Done()
}

func CleanSessionDir(base string, torrents []qbittorrent.Torrent, files []string, wg *sync.WaitGroup) {
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

	wg.Done()
}

func NotifyCrosseed(hash, urlStr string) error {
	if len(urlStr) == 0 {
		return nil
	}

	u := url.Values{}
	u.Set("infoHash", hash)

	client := &http.Client{Timeout: 15 * time.Second}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(u.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}
