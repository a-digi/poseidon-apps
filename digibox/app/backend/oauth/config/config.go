package config

// Zentrale Konfiguration für OAuth-Provider

const (
	DropboxRedirectURI     = "http://localhost:2014/oauth/dropbox/callback"
	DropboxListFolderURL   = "https://api.dropboxapi.com/2/files/list_folder"
	DropboxDownloadFileURL = "https://content.dropboxapi.com/2/files/download"
	DropboxCreateFolderURL = "https://api.dropboxapi.com/2/files/create_folder_v2"
	DropboxUploadFileURL   = "https://content.dropboxapi.com/2/files/upload"
	DropboxDeleteItemURL   = "https://api.dropboxapi.com/2/files/delete_v2"
)
