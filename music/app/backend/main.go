package main

import (
	"encoding/json"
	"fmt"
	"os"

	"music-plugin/analytics"
	"music-plugin/digitalitem"
	"music-plugin/playerstate"
	"music-plugin/playlist"
)

type envelope struct {
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func main() {
	enc := json.NewEncoder(os.Stdout)

	var params map[string]any
	if err := json.NewDecoder(os.Stdin).Decode(&params); err != nil {
		enc.Encode(envelope{Error: "bad request: " + err.Error()})
		return
	}

	db, err := openDB()
	if err != nil {
		enc.Encode(envelope{Error: "db open: " + err.Error()})
		return
	}
	defer db.Close()

	playlistService, err := playlist.NewPlaylistService(db)
	if err != nil {
		enc.Encode(envelope{Error: err.Error()})
		return
	}

	analyticsService, err := analytics.NewAnalyticsService(db)
	if err != nil {
		enc.Encode(envelope{Error: "analytics init: " + err.Error()})
		return
	}

	playerStateService, err := playerstate.NewPlayerStateService(db)
	if err != nil {
		enc.Encode(envelope{Error: "player_state init: " + err.Error()})
		return
	}

	action, _ := params["action"].(string)
	result, dispatchErr := dispatch(action, params, playlistService, analyticsService, playerStateService)
	if dispatchErr != nil {
		enc.Encode(envelope{Error: dispatchErr.Error()})
		return
	}
	enc.Encode(envelope{Result: result})
}

func dispatch(action string, params map[string]any, ps *playlist.PlaylistService, as *analytics.AnalyticsService, pss *playerstate.PlayerStateService) (any, error) {
	switch action {
	case "init_tables":
		return map[string]bool{"ok": true}, nil
	case "list_playlists":
		return json.RawMessage(ps.ListPlaylists()), nil
	case "create_playlist":
		name, _ := params["name"].(string)
		return json.RawMessage(ps.Create(name)), nil
	case "edit_playlist":
		id, _ := params["id"].(string)
		name, _ := params["name"].(string)
		return json.RawMessage(ps.Edit(id, name)), nil
	case "delete_playlist":
		id, _ := params["id"].(string)
		return json.RawMessage(ps.Delete(id)), nil
	case "get_playlist_by_id":
		id, _ := params["id"].(string)
		return json.RawMessage(ps.GetByID(id)), nil
	case "add_playlist_item":
		playlistID, _ := params["playlistId"].(string)
		itemRaw := params["item"]
		itemBytes, err := json.Marshal(itemRaw)
		if err != nil {
			return nil, err
		}
		var item playlist.DigitalItem
		if err := json.Unmarshal(itemBytes, &item); err != nil {
			return nil, err
		}
		return json.RawMessage(ps.AddItem(playlistID, item)), nil
	case "delete_playlist_item":
		playlistID, _ := params["playlistId"].(string)
		itemID, _ := params["itemId"].(string)
		return json.RawMessage(ps.DeleteItem(playlistID, itemID)), nil
	case "get_audio_data_url":
		path, _ := params["path"].(string)
		if path == "" {
			return nil, fmt.Errorf("path required")
		}
		dataURL, err := digitalitem.GetAudioDataURL(path)
		if err != nil {
			return nil, err
		}
		return map[string]string{"dataUrl": dataURL}, nil
	case "record_play":
		itemID, _ := params["itemId"].(string)
		playlistID, _ := params["playlistId"].(string)
		title, _ := params["title"].(string)
		artist, _ := params["artist"].(string)
		album, _ := params["album"].(string)
		return json.RawMessage(as.Record(itemID, playlistID, title, artist, album)), nil

	case "analytics_overview":
		return json.RawMessage(as.Overview()), nil

	case "get_player_state":
		return json.RawMessage(pss.Get()), nil
	case "save_player_state":
		var state playerstate.PlayerState
		raw, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &state); err != nil {
			return nil, err
		}
		return json.RawMessage(pss.Save(state)), nil

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}
