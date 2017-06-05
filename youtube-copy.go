// youtube-copy reads JSON files from Google Takeout for Youtube playlists and
// subscriptions and using the Youtube Data APIs adds them to your Youtube
// account.  It can be use to copy info from one account to another or as a way
// to restore a backup.
package main

import (
	"auth"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/lazyhacker.com/youtube-copy/auth"
	"github.com/lazyhacker/youtube-copy/yt"
)

type Subscriptions []struct {
	ContentDetails struct {
		ActivityType   string `json:"activityType"`
		NewItemCount   int64  `json:"newItemCount"`
		TotalItemCount int64  `json:"totalItemCount"`
	} `json:"contentDetails"`
	Etag    string `json:"etag"`
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Snippet struct {
		ChannelID   string `json:"channelId"`
		Description string `json:"description"`
		PublishedAt string `json:"publishedAt"`
		ResourceID  struct {
			ChannelID string `json:"channelId"`
			Kind      string `json:"kind"`
		} `json:"resourceId"`
		Thumbnails struct {
			Default struct {
				URL string `json:"url"`
			} `json:"default"`
			High struct {
				URL string `json:"url"`
			} `json:"high"`
			Medium struct {
				URL string `json:"url"`
			} `json:"medium"`
		} `json:"thumbnails"`
		Title string `json:"title"`
	} `json:"snippet"`
}

type Playlist []struct {
	ContentDetails struct {
		VideoID          string `json:"videoId"`
		VideoPublishedAt string `json:"videoPublishedAt"`
	} `json:"contentDetails"`
	Etag    string `json:"etag"`
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Snippet struct {
		ChannelID    string `json:"channelId"`
		ChannelTitle string `json:"channelTitle"`
		Description  string `json:"description"`
		PlaylistID   string `json:"playlistId"`
		Position     int64  `json:"position"`
		PublishedAt  string `json:"publishedAt"`
		ResourceID   struct {
			Kind    string `json:"kind"`
			VideoID string `json:"videoId"`
		} `json:"resourceId"`
		Thumbnails struct {
			Default struct {
				Height int64  `json:"height"`
				URL    string `json:"url"`
				Width  int64  `json:"width"`
			} `json:"default"`
			High struct {
				Height int64  `json:"height"`
				URL    string `json:"url"`
				Width  int64  `json:"width"`
			} `json:"high"`
			Maxres struct {
				Height int64  `json:"height"`
				URL    string `json:"url"`
				Width  int64  `json:"width"`
			} `json:"maxres"`
			Medium struct {
				Height int64  `json:"height"`
				URL    string `json:"url"`
				Width  int64  `json:"width"`
			} `json:"medium"`
			Standard struct {
				Height int64  `json:"height"`
				URL    string `json:"url"`
				Width  int64  `json:"width"`
			} `json:"standard"`
		} `json:"thumbnails"`
		Title string `json:"title"`
	} `json:"snippet"`
	Status struct {
		PrivacyStatus string `json:"privacyStatus"`
	} `json:"status"`
}

var (
	clientSecretsFile = flag.String("secrets", "cs.json", "Client Secrets configuration")
	cacheFile         = flag.String("cache", "request.token", "Token cache file")
	transferType      = flag.String("type", "playlist", "playlist (default) or subscription")
	playlist          = flag.String("listname", "default", "Name of playlist to be created and videos added to")
	file              = flag.String("file", "", "Takeout file")
)

func main() {
	flag.Parse()

	oa := auth.SecretsToken{
		ClientSecretsFile: clientSecretsFile,
		CacheFile:         cacheFile,
	}

	ctx := context.WithValue(context.Background(), auth.CtxKey, oa)
	service, err := auth.Authenticate(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var raw []byte
	raw, err = ioutil.ReadFile(*file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Staring transfer")
	switch *transferType {
	case "subscription":

		var data Subscriptions

		if err := json.Unmarshal(raw, &data); err != nil {
			fmt.Printf("Unable to parse JSON. %v\n", err)
			os.Exit(1)
		}

		for i, v := range data {
			fmt.Printf("%d %v\n", i, v.Snippet.ResourceID.ChannelID)
			yt.Subscribe(service, v.Snippet.ResourceID.ChannelID)
		}

	case "playlist":

		var data Playlist
		if err := json.Unmarshal(raw, &data); err != nil {
			fmt.Printf("Unable to parse JSON. %v\n", err)
			os.Exit(1)
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
