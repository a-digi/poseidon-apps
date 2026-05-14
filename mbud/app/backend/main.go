package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	recurringpkg "mbud-plugin/recurring"
	"mbud-plugin/storage"
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

	dataDir := resolveDataDir()
	db, err := storage.Open(dataDir)
	if err != nil {
		enc.Encode(envelope{Error: "open db: " + err.Error()})
		return
	}
	defer db.Close()

	invRepo := storage.NewInvoiceRepo(db)
	recRepo := storage.NewRecurringRepo(db)
	upcomingRepo := storage.NewUpcomingRepo(db)
	links := storage.NewLinksRepo(db)
	upcomingLinks := storage.NewUpcomingLinksRepo(db)
	userRepo := storage.NewUserRepo(db)
	userLinks := storage.NewUserLinksRepo(db)
	recurringUserLinks := storage.NewRecurringUserLinksRepo(db)
	upcomingUserLinks := storage.NewUpcomingUserLinksRepo(db)
	tagRepo := storage.NewTagRepo(db)
	businessTagsRepo := storage.NewBusinessTagsRepo(db)
	invoiceTagsRepo := storage.NewInvoiceTagsRepo(db)
	recurringTagsRepo := storage.NewRecurringTagsRepo(db)
	upcomingTagsRepo := storage.NewUpcomingTagsRepo(db)
	reconciler := recurringpkg.New(db, recRepo, upcomingRepo, invRepo, links, upcomingLinks, newID, func() int64 { return time.Now().Unix() })
	uploadSessionRepo := storage.NewUploadSessionRepo(db)
	attachmentRepo := storage.NewInvoiceAttachmentRepo(db)

	h := &handlers{
		dataDir:            dataDir,
		db:                 db,
		businesses:         storage.NewBusinessRepo(db),
		invoices:           invRepo,
		recurring:          recRepo,
		upcoming:           upcomingRepo,
		links:              links,
		upcomingLinks:      upcomingLinks,
		users:              userRepo,
		userLinks:          userLinks,
		recurringUserLinks: recurringUserLinks,
		upcomingUserLinks:  upcomingUserLinks,
		tags:               tagRepo,
		businessTags:       businessTagsRepo,
		invoiceTags:        invoiceTagsRepo,
		recurringTags:      recurringTagsRepo,
		upcomingTags:       upcomingTagsRepo,
		reconciler:         reconciler,
		uploadSessions:     uploadSessionRepo,
		attachments:        attachmentRepo,
	}

	action, _ := params["action"].(string)
	result, err := h.dispatch(context.Background(), action, params)
	if err != nil {
		enc.Encode(envelope{Error: err.Error()})
		return
	}
	enc.Encode(envelope{Result: result})
}

func resolveDataDir() string {
	if dir := os.Getenv("PLUGIN_DATA_DIR"); dir != "" {
		return dir
	}
	exe, err := os.Executable()
	if err != nil {
		return "./data"
	}
	pluginRoot := filepath.Dir(filepath.Dir(exe))
	return filepath.Join(pluginRoot, "data")
}
