package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	qbittorrent "github.com/autobrr/go-qbittorrent"
)

func main() {
	var password, host, username, path string
	flag.StringVar(&username, "U", os.Getenv("USERNAME"), "username")
	flag.StringVar(&host, "H", os.Getenv("HOST"), "host")
	flag.StringVar(&password, "P", os.Getenv("PASSWORD"), "password")
	flag.StringVar(&path, "S", os.Getenv("SESSIONDIR"), "BT_backup")

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
	}
}
