package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"digibox-plugin/downloads"
	"digibox-plugin/dropbox"
	"digibox-plugin/oauth"
	"digibox-plugin/oauth/builtin"
	oauthdropbox "digibox-plugin/oauth/dropbox"
	"digibox-plugin/oauth/model"
	"digibox-plugin/oauth/repository"
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

	builtins, err := builtin.Load()
	if err != nil {
		enc.Encode(envelope{Error: err.Error()})
		return
	}
	clientRepo, err := repository.NewOauthClientsRepository(db, builtins)
	if err != nil {
		enc.Encode(envelope{Error: err.Error()})
		return
	}
	tokenRepo, err := repository.NewOauthTokenRepository(db)
	if err != nil {
		enc.Encode(envelope{Error: err.Error()})
		return
	}
	requestRepo, err := repository.NewOauthRequestRepository(db)
	if err != nil {
		enc.Encode(envelope{Error: err.Error()})
		return
	}
	profileRepo, err := repository.NewOauthProfileRepository(db)
	if err != nil {
		enc.Encode(envelope{Error: err.Error()})
		return
	}
	downloadsRepo, err := downloads.NewDownloadsRepository(db)
	if err != nil {
		enc.Encode(envelope{Error: err.Error()})
		return
	}

	oauth.InitProviderChain(tokenRepo)

	action, _ := params["action"].(string)
	result, dispatchErr := dispatch(action, params, db, clientRepo, tokenRepo, requestRepo, profileRepo, downloadsRepo)
	if dispatchErr != nil {
		enc.Encode(envelope{Error: dispatchErr.Error()})
		return
	}
	enc.Encode(envelope{Result: result})
}

func dispatch(
	action string,
	params map[string]any,
	db *sql.DB,
	clientRepo *repository.OauthClientsRepository,
	tokenRepo *repository.OauthTokenRepository,
	requestRepo *repository.OauthRequestRepository,
	profileRepo *repository.OauthProfileRepository,
	downloadsRepo *downloads.DownloadsRepository,
) (any, error) {
	switch action {
	case "init_tables":
		return map[string]bool{"ok": true}, nil
	case "list_clients":
		return unwrap(oauth.ListOAuthClients(clientRepo))
	case "add_client":
		clientJSON, err := paramsAsClientJSON(params)
		if err != nil {
			return nil, err
		}
		return unwrap(oauth.AddOAuthClient(clientRepo, clientJSON))
	case "delete_client":
		id, _ := params["id"].(string)
		if id == "" {
			return nil, fmt.Errorf("id required")
		}
		if _, err := unwrap(oauth.DeleteOAuthClient(clientRepo, id)); err != nil {
			return nil, err
		}
		return map[string]bool{"ok": true}, nil
	case "get_active_providers":
		return unwrap(oauth.GetActiveProviders(clientRepo))
	case "create_auth_link":
		clientID, _ := params["clientId"].(string)
		if clientID == "" {
			return nil, fmt.Errorf("clientId required")
		}
		state, _ := params["state"].(string)
		if state == "" {
			state = uuid.NewString()
		}
		return unwrap(oauthdropbox.CreateAuthorizationLink(clientRepo, requestRepo, clientID, state))
	case "exchange_code":
		state, _ := params["state"].(string)
		code, _ := params["code"].(string)
		return unwrap(oauthdropbox.OauthCodeAuthorize(requestRepo, clientRepo, tokenRepo, profileRepo, state, code))
	case "get_oauth_code":
		return getOAuthCode()
	case "list_authorizations":
		return listAuthorizations(clientRepo, tokenRepo, requestRepo, profileRepo)
	case "delete_authorization":
		id, _ := params["id"].(string)
		if id == "" {
			return nil, fmt.Errorf("id required")
		}
		return deleteAuthorization(db, tokenRepo, id)
	case "list_files":
		return listFiles(params, clientRepo, tokenRepo, downloadsRepo)
	case "download_file":
		return downloadFile(params, tokenRepo, downloadsRepo)
	case "create_folder":
		return createFolder(params, clientRepo, tokenRepo)
	case "upload_file":
		return uploadFile(params, clientRepo, tokenRepo)
	case "delete_item":
		return deleteItem(params, clientRepo, tokenRepo)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func unwrap(raw string) (any, error) {
	var resp struct {
		Status  string `json:"status"`
		Message any    `json:"message"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, fmt.Errorf("invalid response: %s", raw)
	}
	if resp.Status == "success" {
		return resp.Message, nil
	}
	if msg, ok := resp.Message.(string); ok {
		return nil, fmt.Errorf("%s", msg)
	}
	return nil, fmt.Errorf("%v", resp.Message)
}

func paramsAsClientJSON(params map[string]any) (string, error) {
	cleaned := make(map[string]any, len(params))
	for k, v := range params {
		if k == "action" {
			continue
		}
		cleaned[k] = v
	}
	b, err := json.Marshal(cleaned)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getOAuthCode() (any, error) {
	path := filepath.Join(os.TempDir(), "digibox_oauth.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{"pending": true}, nil
		}
		return nil, err
	}
	os.Remove(path)
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func listAuthorizations(
	clientRepo *repository.OauthClientsRepository,
	tokenRepo *repository.OauthTokenRepository,
	requestRepo *repository.OauthRequestRepository,
	profileRepo *repository.OauthProfileRepository,
) (any, error) {
	tokens, err := tokenRepo.FindAll()
	if err != nil {
		return nil, err
	}
	views := []model.OauthRequestView{}
	for _, token := range tokens {
		requestedOn := int64(0)
		if token.RequestID != "" {
			req, err := requestRepo.FindById(token.RequestID)
			if err != nil || req == nil {
				log.Printf("[listAuthorizations] no request for token %s (request_id=%s): %v", token.ID, token.RequestID, err)
			} else {
				requestedOn = req.RequestedOn
			}
		}
		var profile *model.OauthProfile = nil
		if token.ID != "" && profileRepo != nil {
			if p, err := profileRepo.FindByTokenId(token.ID); err == nil && p != nil {
				profile = p
			}
		}
		if c, err := clientRepo.FindByClientId(token.ClientId); err == nil {
			views = append(views, model.OauthRequestView{
				ID:            token.RequestID,
				RequestedOn:   requestedOn,
				OauthClientID: c.ID,
				OauthToken:    token,
				OauthProfile:  profile,
			})
		} else {
			views = append(views, model.OauthRequestView{
				ID:           token.RequestID,
				RequestedOn:  requestedOn,
				OauthToken:   token,
				OauthProfile: profile,
			})
		}
	}
	return views, nil
}

func deleteAuthorization(db *sql.DB, tokenRepo *repository.OauthTokenRepository, requestID string) (any, error) {
	token, _ := tokenRepo.FindByTokenId(requestID)
	if token != nil {
		if err := tokenRepo.DeleteByID(token.ID); err != nil {
			return nil, err
		}
		return map[string]bool{"ok": true}, nil
	}
	if _, err := db.Exec(`DELETE FROM oauth_requests WHERE id = ?`, requestID); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}

func listFiles(
	params map[string]any,
	clientRepo *repository.OauthClientsRepository,
	tokenRepo *repository.OauthTokenRepository,
	downloadsRepo *downloads.DownloadsRepository,
) (any, error) {
	tokenID, _ := params["tokenId"].(string)
	path, _ := params["path"].(string)
	if tokenID == "" {
		return nil, fmt.Errorf("tokenId required")
	}

	token, err := tokenRepo.FindByID(tokenID)
	if err != nil {
		return nil, err
	}
	client, err := clientRepo.FindByClientId(token.ClientId)
	if err != nil {
		return nil, err
	}
	result, err := dropbox.ListDropboxFiles(token, tokenRepo, client.Secret, path)
	if err != nil {
		return nil, err
	}
	for i := range result.Entries {
		found, _ := downloadsRepo.FindByPlatformAndExternalId("dropbox", result.Entries[i].Path)
		if found != nil {
			result.Entries[i].Downloaded = true
			result.Entries[i].TargetFolder = found.TargetFolder
		}
	}
	return result, nil
}

func downloadFile(
	params map[string]any,
	tokenRepo *repository.OauthTokenRepository,
	downloadsRepo *downloads.DownloadsRepository,
) (any, error) {
	tokenID, _ := params["tokenId"].(string)
	path, _ := params["path"].(string)
	targetDir, _ := params["targetDir"].(string)
	if tokenID == "" || path == "" || targetDir == "" {
		return nil, fmt.Errorf("tokenId, path and targetDir are required")
	}

	token, err := tokenRepo.FindByID(tokenID)
	if err != nil {
		return nil, err
	}

	body, err := dropbox.DownloadDropboxFile(token.AccessToken, path)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	filename := filepath.Base(path)
	if filename == "" || filename == "/" || filename == "." {
		return nil, fmt.Errorf("invalid path: %s", path)
	}
	fullPath := filepath.Join(targetDir, filename)
	if err := dropbox.DownloadFromBody(body, fullPath); err != nil {
		return nil, err
	}

	if err := downloadsRepo.Insert("dropbox", path, targetDir); err != nil {
		log.Printf("[downloadFile] insert failed: %v", err)
	}

	return map[string]any{"ok": true, "path": fullPath}, nil
}

func createFolder(
	params map[string]any,
	clientRepo *repository.OauthClientsRepository,
	tokenRepo *repository.OauthTokenRepository,
) (any, error) {
	tokenID, _ := params["tokenId"].(string)
	path, _ := params["path"].(string)
	if tokenID == "" || path == "" {
		return nil, fmt.Errorf("tokenId and path are required")
	}
	token, err := tokenRepo.FindByID(tokenID)
	if err != nil {
		return nil, err
	}
	client, err := clientRepo.FindByClientId(token.ClientId)
	if err != nil {
		return nil, err
	}
	if err := dropbox.CreateDropboxFolder(token, tokenRepo, client.Secret, path); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

func uploadFile(
	params map[string]any,
	clientRepo *repository.OauthClientsRepository,
	tokenRepo *repository.OauthTokenRepository,
) (any, error) {
	tokenID, _ := params["tokenId"].(string)
	path, _ := params["path"].(string)
	sourcePath, _ := params["sourcePath"].(string)
	if tokenID == "" || path == "" || sourcePath == "" {
		return nil, fmt.Errorf("tokenId, path and sourcePath are required")
	}
	token, err := tokenRepo.FindByID(tokenID)
	if err != nil {
		return nil, err
	}
	client, err := clientRepo.FindByClientId(token.ClientId)
	if err != nil {
		return nil, err
	}
	if err := dropbox.UploadDropboxFile(token, tokenRepo, client.Secret, sourcePath, path); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "path": path}, nil
}

func deleteItem(
	params map[string]any,
	clientRepo *repository.OauthClientsRepository,
	tokenRepo *repository.OauthTokenRepository,
) (any, error) {
	tokenID, _ := params["tokenId"].(string)
	path, _ := params["path"].(string)
	if tokenID == "" || path == "" {
		return nil, fmt.Errorf("tokenId and path are required")
	}
	token, err := tokenRepo.FindByID(tokenID)
	if err != nil {
		return nil, err
	}
	client, err := clientRepo.FindByClientId(token.ClientId)
	if err != nil {
		return nil, err
	}
	if err := dropbox.DeleteDropboxItem(token, tokenRepo, client.Secret, path); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}
