// Package yt is a wrapper around the Youtube Data API for creating new
// playlists, add videos to a playlist and subscribing to channels.
package yt // import "lazyhacker.dev/youtube-copy/yt"

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	youtube "google.golang.org/api/youtube/v3"
)

func handleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}

func addPropertyToResource(ref map[string]interface{}, keys []string, value string, count int) map[string]interface{} {
	for k := count; k < (len(keys) - 1); k++ {
		switch val := ref[keys[k]].(type) {
		case map[string]interface{}:
			ref[keys[k]] = addPropertyToResource(val, keys, value, (k + 1))
		case nil:
			next := make(map[string]interface{})
			ref[keys[k]] = addPropertyToResource(next, keys, value, (k + 1))
		}
	}
	// Only include properties that have values.
	if count == len(keys)-1 && value != "" {
		valueKey := keys[len(keys)-1]
		if valueKey[len(valueKey)-2:] == "[]" {
			ref[valueKey[0:len(valueKey)-2]] = strings.Split(value, ",")
		} else if len(valueKey) > 4 && valueKey[len(valueKey)-4:] == "|int" {
			ref[valueKey[0:len(valueKey)-4]], _ = strconv.Atoi(value)
		} else if value == "true" {
			ref[valueKey] = true
		} else if value == "false" {
			ref[valueKey] = false
		} else {
			ref[valueKey] = value
		}
	}
	return ref
}

func createResource(properties map[string]string) string {
	resource := make(map[string]interface{})
	for key, value := range properties {
		keys := strings.Split(key, ".")
		ref := addPropertyToResource(resource, keys, value, 0)
		resource = ref
	}
	propJson, err := json.Marshal(resource)
	if err != nil {
		log.Fatal("cannot encode to JSON ", err)
	}
	return string(propJson)
}
func printPlaylistsListResults(response *youtube.PlaylistListResponse) {
	for _, item := range response.Items {
		fmt.Println(item.Id, ": ", item.Snippet.Title)
	}
}

func playlistsListByChannelId(service *youtube.Service, part string, channelId string, maxResults int64) {
	call := service.Playlists.List(part)
	if channelId != "" {
		call = call.ChannelId(channelId)
	}
	if maxResults != 0 {
		call = call.MaxResults(maxResults)
	}
	response, err := call.Do()
	handleError(err, "")
	printPlaylistsListResults(response)
}

func printPlaylistItemsInsertResults(response *youtube.PlaylistItem) {
	// Handle response here
}

func playlistItemsInsert(service *youtube.Service, part string, onBehalfOfContentOwner string, res string) {
	resource := &youtube.PlaylistItem{}
	if err := json.NewDecoder(strings.NewReader(res)).Decode(&resource); err != nil {
		log.Fatal(err)
	}
	call := service.PlaylistItems.Insert(part, resource)
	response, err := call.Do()
	if err != nil {
		log.Printf("Error adding video to playlist. %v", err)
	}
	printPlaylistItemsInsertResults(response)
}

func printSubscriptionsInsertResults(response *youtube.Subscription) {
	// Handle response here
}

func subscriptionsInsert(service *youtube.Service, part string, res string) {
	resource := &youtube.Subscription{}
	if err := json.NewDecoder(strings.NewReader(res)).Decode(&resource); err != nil {
		log.Fatal(err)
	}
	call := service.Subscriptions.Insert(part, resource)
	response, err := call.Do()
	handleError(err, "")
	printSubscriptionsInsertResults(response)
}

func printPlaylistsInsertResults(response *youtube.Playlist) {
	// Handle response here
}

func playlistsInsert(service *youtube.Service, part string, onBehalfOfContentOwner string, res string) *youtube.Playlist {
	resource := &youtube.Playlist{}
	if err := json.NewDecoder(strings.NewReader(res)).Decode(&resource); err != nil {
		log.Fatal(err)
	}
	call := service.Playlists.Insert(part, resource)
	response, err := call.Do()

	if err != nil {
		log.Fatalf("Error creating playlist.  %v", err)
	}
	//printPlaylistsInsertResults(response)
	return response
}

func CreatePlaylist(service *youtube.Service, title, description, privacy string) string {
	// Create new playlist
	properties := (map[string]string{
		"snippet.title":           title,
		"snippet.description":     description,
		"snippet.tags[]":          "",
		"snippet.defaultLanguage": "",
		"status.privacyStatus":    privacy,
	})
	res := createResource(properties)
	response := playlistsInsert(service, "snippet,status", "", res)
	return response.Id
}

func AddVideoToPlaylist(service *youtube.Service, playlist, video string) {
	// insert videos into playlist
	properties := (map[string]string{
		"snippet.playlistId":         playlist,
		"snippet.resourceId.kind":    "youtube#video",
		"snippet.resourceId.videoId": video,
		"snippet.position":           "",
	})
	res := createResource(properties)
	playlistItemsInsert(service, "snippet", "", res)
}

func Subscribe(service *youtube.Service, channel string) {
	// add subscriptions
	properties := (map[string]string{
		"snippet.resourceId.kind":      "youtube#channel",
		"snippet.resourceId.channelId": channel,
	})
	res := createResource(properties)
	subscriptionsInsert(service, "snippet", res)
}

func GetPlaylists(service *youtube.Service, channel string) {

	playlistsListByChannelId(service, "snippet,contentDetails", channel, 25)
}
