// youtube-copy reads JSON files from Google Takeout for Youtube playlists and
// subscriptions and using the Youtube Data APIs adds them to your Youtube
// account.  It can be use to copy info from one account to another or as a way
// to restore a backup.
package main // import "lazyhacker.dev/youtube-copy"

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2"
	youtube "google.golang.org/api/youtube/v3"
	"lazyhacker.dev/gclientauth"
	"lazyhacker.dev/youtube-copy/internal/yt"
)

var (
	clientSecretsFile = flag.String("secrets", "cs.json", "Client Secrets configuration")
	cacheFile         = flag.String("cache", "request.token", "Token cache file")
	transferType      = flag.String("type", "playlist", "playlist (default) or subscription")
	playlist          = flag.String("listname", "default", "Name of playlist to be created and videos added to")
	file              = flag.String("file", "", "Takeout file")
)

func main() {
	flag.Parse()

	scopes := []string{youtube.YoutubeForceSslScope}

	ctx := oauth2.NoContext
	token, config, err := gclientauth.GetGoogleOauth2Token(ctx, *clientSecretsFile, *cacheFile, scopes, false, "8081")
	if err != nil {
		log.Fatalf("Fetching oAuth token failed.\n%v", err)
	}
	cfg := config.Client(ctx, token)
	defer cfg.CloseIdleConnections()
	if err != nil {
		log.Fatal(err)
	}

	service, err := youtube.New(cfg)

	if err != nil {
		log.Fatalln(err)
	}

	var raw []byte
	raw, err = ioutil.ReadFile(*file)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Starting transfer")
	switch *transferType {
	case "subscription":

		var data yt.Subscriptions

		if err := json.Unmarshal(raw, &data); err != nil {
			log.Fatalf("Unable to parse JSON. %v\n", err)
		}

		for i, v := range data {
			log.Printf("%d %v\n", i, v.Snippet.ResourceID.ChannelID)
			yt.Subscribe(service, v.Snippet.ResourceID.ChannelID)
		}

	case "playlist":

		var data yt.Playlist
		if err := json.Unmarshal(raw, &data); err != nil {
			log.Fatalf("Unable to parse JSON. %v\n", err)
		}

		playlistID := yt.CreatePlaylist(service, *playlist, "", "private")
		fmt.Printf("New Playlist ID = %v\n", playlistID)

		for i, v := range data {
			fmt.Printf("%d %v\n", i, v.ContentDetails.VideoID)
			yt.AddVideoToPlaylist(service, playlistID, v.ContentDetails.VideoID)
		}
	default:
		fmt.Println("Unknow type.")
	}
}
