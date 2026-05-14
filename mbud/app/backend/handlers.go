package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"mbud-plugin/model"
	recurringpkg "mbud-plugin/recurring"
	"mbud-plugin/storage"
)

const defaultCurrency = "EUR"

type handlers struct {
	dataDir            string
	db                 *sql.DB
	businesses         *storage.BusinessRepo
	invoices           *storage.InvoiceRepo
	recurring          *storage.RecurringRepo
	upcoming           *storage.UpcomingRepo
	links              *storage.LinksRepo
	upcomingLinks      *storage.UpcomingLinksRepo
	users              *storage.UserRepo
	userLinks          *storage.UserLinksRepo
	recurringUserLinks *storage.RecurringUserLinksRepo
	upcomingUserLinks  *storage.UpcomingUserLinksRepo
	tags               *storage.TagRepo
	businessTags       *storage.BusinessTagsRepo
	invoiceTags        *storage.InvoiceTagsRepo
	recurringTags      *storage.RecurringTagsRepo
	upcomingTags       *storage.UpcomingTagsRepo
	reconciler         *recurringpkg.Reconciler
	uploadSessions     *storage.UploadSessionRepo
	attachments        *storage.InvoiceAttachmentRepo
}

const maxLogoBytes = 1024 * 1024

var allowedInvoiceSortColumns = map[string]struct{}{
	"issuedAt": {},
	"dueAt":    {},
	"amount":   {},
}

func extFromMime(mime string) string {
	switch mime {
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	case "application/pdf":
		return "pdf"
	}
	return ""
}

func businessLogoPath(dataDir, id, mime string) string {
	return filepath.Join(dataDir, "businesses", id+"."+extFromMime(mime))
}

func (h *handlers) attachLogo(b *model.Business) {
	if b.LogoType == "" {
		return
	}
	bytes, err := os.ReadFile(businessLogoPath(h.dataDir, b.ID, b.LogoType))
	if err != nil {
		return
	}
	b.Logo = "data:" + b.LogoType + ";base64," + base64.StdEncoding.EncodeToString(bytes)
}

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func nowUnix() int64 { return time.Now().Unix() }

func decodeInto(params map[string]any, key string, out any) error {
	raw, err := json.Marshal(params[key])
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}

func decodeStringArray(m map[string]any, key string) []string {
	raw, ok := m[key].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func validFrequency(f model.Frequency) bool {
	switch f {
	case model.FrequencyDaily, model.FrequencyWeekly, model.FrequencyMonthly, model.FrequencyYearly:
		return true
	}
	return false
}

func validateAnchors(r model.RecurringInvoice) error {
	switch r.Frequency {
	case model.FrequencyWeekly:
		if r.IssueDayOfWeek < 1 || r.IssueDayOfWeek > 7 {
			return errors.New("issueDayOfWeek must be 1 (Monday) through 7 (Sunday)")
		}
	case model.FrequencyMonthly:
		if r.IssueDayOfMonth < 1 || r.IssueDayOfMonth > 31 {
			return errors.New("issueDayOfMonth must be 1–31")
		}
	case model.FrequencyYearly:
		if r.IssueMonthOfYear < 1 || r.IssueMonthOfYear > 12 {
			return errors.New("issueMonthOfYear must be 1–12")
		}
		if r.IssueDayOfMonth < 1 || r.IssueDayOfMonth > 31 {
			return errors.New("issueDayOfMonth must be 1–31")
		}
	}
	return nil
}

func (h *handlers) dispatch(ctx context.Context, action string, params map[string]any) (any, error) {
	switch action {
	case "health":
		return map[string]any{"ok": true, "version": "0.1.0"}, nil

	case "list_business":
		return h.listBusiness(ctx)
	case "create_business":
		return h.createBusiness(ctx, params)
	case "update_business":
		return h.updateBusiness(ctx, params)
	case "delete_business":
		return h.deleteBusiness(ctx, params)
	case "upload_business_logo":
		return h.uploadBusinessLogo(ctx, params)
	case "delete_business_logo":
		return h.deleteBusinessLogo(ctx, params)

	case "list_invoice":
		return h.listInvoice(ctx, params)
	case "invoice_stats":
		return h.invoiceStats(ctx, params)
	case "create_invoice":
		return h.createInvoice(ctx, params)
	case "update_invoice":
		return h.updateInvoice(ctx, params)
	case "delete_invoice":
		return h.deleteInvoice(ctx, params)

	case "list_recurring":
		return h.listRecurring(ctx)
	case "create_recurring":
		return h.createRecurring(ctx, params)
	case "update_recurring":
		return h.updateRecurring(ctx, params)
	case "delete_recurring":
		return h.deleteRecurring(ctx, params)

	case "list_upcoming":
		return h.listUpcoming(ctx)
	case "create_upcoming":
		return h.createUpcoming(ctx, params)
	case "update_upcoming":
		return h.updateUpcoming(ctx, params)
	case "delete_upcoming":
		return h.deleteUpcoming(ctx, params)

	case "list_user":
		return h.listUser(ctx)
	case "create_user":
		return h.createUser(ctx, params)
	case "update_user":
		return h.updateUser(ctx, params)
	case "delete_user":
		return h.deleteUser(ctx, params)

	case "list_tag":
		return h.listTag(ctx)
	case "create_tag":
		return h.createTag(ctx, params)
	case "update_tag":
		return h.updateTag(ctx, params)
	case "delete_tag":
		return h.deleteTag(ctx, params)

	case "create_upload_session":
		return h.createUploadSession(ctx, params)
	case "get_upload_session":
		return h.getUploadSession(ctx, params)
	case "mobile_upload_file":
		return h.mobileUploadFile(ctx, params)
	case "list_invoice_attachments":
		return h.listInvoiceAttachments(ctx, params)
	case "attach_session_to_invoice":
		return h.attachSessionToInvoice(ctx, params)
	case "delete_attachment":
		return h.deleteAttachment(ctx, params)
	case "get_attachment_bytes":
		return h.getAttachmentBytes(ctx, params)

	default:
		return nil, errors.New("unknown action: " + action)
	}
}

// ── Business ──────────────────────────────────────────────────────────────────

func (h *handlers) listBusiness(ctx context.Context) (any, error) {
	list, err := h.businesses.List(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(list))
	for i := range list {
		ids[i] = list[i].ID
	}
	tagMapping, err := h.businessTags.TagIDsBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i].TagIDs = tagMapping[list[i].ID]
		h.attachLogo(&list[i])
	}
	return list, nil
}

func (h *handlers) createBusiness(ctx context.Context, params map[string]any) (any, error) {
	var b model.Business
	if err := decodeInto(params, "business", &b); err != nil {
		return nil, err
	}
	b.Name = strings.TrimSpace(b.Name)
	if b.Name == "" {
		return nil, errors.New("name is required")
	}
	b.LogoType = ""
	b.Logo = ""
	now := nowUnix()
	b.ID = newID()
	b.CreatedAt = now
	b.UpdatedAt = now
	if err := h.businesses.Insert(ctx, b); err != nil {
		return nil, err
	}
	validatedTags, err := h.validateTagIDs(ctx, b.TagIDs)
	if err != nil {
		return nil, err
	}
	if err := h.businessTags.ReplaceForBusiness(ctx, b.ID, validatedTags); err != nil {
		return nil, err
	}
	b.TagIDs = validatedTags
	h.attachLogo(&b)
	return b, nil
}

func (h *handlers) updateBusiness(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	var b model.Business
	if err := decodeInto(params, "business", &b); err != nil {
		return nil, err
	}
	b.Name = strings.TrimSpace(b.Name)
	if b.Name == "" {
		return nil, errors.New("name is required")
	}
	existing, ok, err := h.businesses.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("business not found")
	}
	b.ID = id
	b.CreatedAt = existing.CreatedAt
	b.UpdatedAt = nowUnix()
	b.LogoType = existing.LogoType
	b.Logo = ""
	if _, err := h.businesses.Update(ctx, b); err != nil {
		return nil, err
	}
	validatedTags, err := h.validateTagIDs(ctx, b.TagIDs)
	if err != nil {
		return nil, err
	}
	if err := h.businessTags.ReplaceForBusiness(ctx, b.ID, validatedTags); err != nil {
		return nil, err
	}
	b.TagIDs = validatedTags
	h.attachLogo(&b)
	return b, nil
}

func (h *handlers) deleteBusiness(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	existing, ok, err := h.businesses.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := h.businesses.Delete(ctx, id); err != nil {
		return nil, err
	}
	if ok && existing.LogoType != "" {
		if err := os.Remove(businessLogoPath(h.dataDir, id, existing.LogoType)); err != nil && !os.IsNotExist(err) {
			_ = err
		}
	}
	return map[string]any{"ok": true}, nil
}

func (h *handlers) uploadBusinessLogo(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	if id == "" {
		return nil, errors.New("id is required")
	}
	dataBase64, _ := params["dataBase64"].(string)
	if dataBase64 == "" {
		return nil, errors.New("dataBase64 is required")
	}
	contentType, _ := params["contentType"].(string)
	if contentType == "" {
		return nil, errors.New("contentType is required")
	}
	if extFromMime(contentType) == "" {
		return nil, errors.New("unsupported logo type")
	}
	decoded, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return nil, errors.New("invalid base64")
	}
	if len(decoded) > maxLogoBytes {
		return nil, errors.New("logo too large (max 1 MB)")
	}
	existing, ok, err := h.businesses.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("business not found")
	}
	if existing.LogoType != "" && extFromMime(existing.LogoType) != extFromMime(contentType) {
		if err := os.Remove(businessLogoPath(h.dataDir, id, existing.LogoType)); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	if err := os.MkdirAll(filepath.Join(h.dataDir, "businesses"), 0o755); err != nil {
		return nil, err
	}
	finalPath := businessLogoPath(h.dataDir, id, contentType)
	tmpPath := finalPath + ".tmp"
	if err := os.WriteFile(tmpPath, decoded, 0o644); err != nil {
		return nil, err
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return nil, err
	}
	now := nowUnix()
	updated, err := h.businesses.UpdateLogoType(ctx, id, contentType, now)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, errors.New("business not found")
	}
	existing.LogoType = contentType
	existing.UpdatedAt = now
	existing.Logo = "data:" + contentType + ";base64," + dataBase64
	return existing, nil
}

func (h *handlers) deleteBusinessLogo(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	if id == "" {
		return nil, errors.New("id is required")
	}
	existing, ok, err := h.businesses.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("business not found")
	}
	if existing.LogoType == "" {
		return existing, nil
	}
	if err := os.Remove(businessLogoPath(h.dataDir, id, existing.LogoType)); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	now := nowUnix()
	if _, err := h.businesses.UpdateLogoType(ctx, id, "", now); err != nil {
		return nil, err
	}
	existing.LogoType = ""
	existing.Logo = ""
	existing.UpdatedAt = now
	return existing, nil
}

// ── Invoice ───────────────────────────────────────────────────────────────────

type invoiceListResult struct {
	Items                []model.Invoice        `json:"items"`
	Total                int                    `json:"total"`
	AvailableBusinessIDs []string               `json:"availableBusinessIds"`
	AvailableUserIDs     []string               `json:"availableUserIds"`
	AvailableTagIDs      []string               `json:"availableTagIds"`
	PendingItems         []model.PendingInvoice `json:"pendingItems"`
}

func (h *handlers) listInvoice(ctx context.Context, params map[string]any) (any, error) {
	fromF, _ := params["from"].(float64)
	toF, _ := params["to"].(float64)
	limitF, _ := params["limit"].(float64)
	offsetF, _ := params["offset"].(float64)
	from := int64(fromF)
	to := int64(toF)
	limit := int64(limitF)
	offset := int64(offsetF)
	if from < 0 {
		return nil, errors.New("from must be ≥ 0")
	}
	if to < 0 {
		return nil, errors.New("to must be ≥ 0")
	}
	if limit < 0 {
		return nil, errors.New("limit must be ≥ 0")
	}
	if offset < 0 {
		return nil, errors.New("offset must be ≥ 0")
	}
	if from > 0 && to > 0 && from > to {
		return nil, errors.New("from must be ≤ to")
	}
	sortBy, _ := params["sortBy"].(string)
	sortDir, _ := params["sortDir"].(string)
	if sortBy == "" {
		sortBy = "issuedAt"
	}
	if sortDir == "" {
		sortDir = "desc"
	}
	if _, ok := allowedInvoiceSortColumns[sortBy]; !ok {
		return nil, errors.New("sortBy must be one of: issuedAt, dueAt, amount")
	}
	if sortDir != "asc" && sortDir != "desc" {
		return nil, errors.New("sortDir must be 'asc' or 'desc'")
	}
	businessIDs := decodeStringArray(params, "businessIds")
	userIDs := decodeStringArray(params, "userIds")
	tagIDs := decodeStringArray(params, "tagIds")
	unpaidOnly, _ := params["unpaidOnly"].(bool)
	if err := h.reconciler.Catchup(ctx); err != nil {
		return nil, err
	}
	availableBusinessIDs, err := h.invoices.DistinctBusinessIDs(ctx, from, to)
	if err != nil {
		return nil, err
	}
	availableUserIDs, err := h.userLinks.DistinctUserIDs(ctx, from, to)
	if err != nil {
		return nil, err
	}
	availableTagIDs, err := h.invoiceTags.DistinctTagIDs(ctx, from, to)
	if err != nil {
		return nil, err
	}
	total, err := h.invoices.Count(ctx, from, to, businessIDs, userIDs, tagIDs, unpaidOnly)
	if err != nil {
		return nil, err
	}
	items, err := h.invoices.List(ctx, from, to, businessIDs, userIDs, tagIDs, unpaidOnly, limit, offset, sortBy, sortDir)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	mapping, err := h.links.RecurringIDsBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if rids, ok := mapping[items[i].ID]; ok {
			items[i].RecurringIDs = rids
		}
	}
	upMapping, err := h.upcomingLinks.UpcomingIDsBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if uids, ok := upMapping[items[i].ID]; ok {
			items[i].UpcomingIDs = uids
		}
	}
	userMapping, err := h.userLinks.UserIDsBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if uids, ok := userMapping[items[i].ID]; ok {
			items[i].UserIDs = uids
		}
	}
	tagMapping, err := h.invoiceTags.TagIDsBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if tids, ok := tagMapping[items[i].ID]; ok {
			items[i].TagIDs = tids
		}
	}
	pending, err := recurringpkg.ProjectPending(ctx, h.db, nowUnix(), recurringpkg.Filters{
		From:        from,
		To:          to,
		BusinessIDs: businessIDs,
		UserIDs:     userIDs,
		TagIDs:      tagIDs,
	})
	if err != nil {
		return nil, err
	}
	return invoiceListResult{Items: items, Total: total, AvailableBusinessIDs: availableBusinessIDs, AvailableUserIDs: availableUserIDs, AvailableTagIDs: availableTagIDs, PendingItems: pending}, nil
}

func (h *handlers) invoiceStats(ctx context.Context, params map[string]any) (any, error) {
	fromF, _ := params["from"].(float64)
	toF, _ := params["to"].(float64)
	from := int64(fromF)
	to := int64(toF)
	if from < 0 {
		return nil, errors.New("from must be ≥ 0")
	}
	if to < 0 {
		return nil, errors.New("to must be ≥ 0")
	}
	if from > 0 && to > 0 && from > to {
		return nil, errors.New("from must be ≤ to")
	}
	businessIDs := decodeStringArray(params, "businessIds")
	userIDs := decodeStringArray(params, "userIds")
	tagIDs := decodeStringArray(params, "tagIds")
	if err := h.reconciler.Catchup(ctx); err != nil {
		return nil, err
	}
	stats, err := h.invoices.Stats(ctx, from, to, businessIDs, userIDs, tagIDs)
	if err != nil {
		return nil, err
	}
	pending, err := recurringpkg.ProjectPending(ctx, h.db, nowUnix(), recurringpkg.Filters{
		From:        from,
		To:          to,
		BusinessIDs: businessIDs,
		UserIDs:     userIDs,
		TagIDs:      tagIDs,
	})
	if err != nil {
		return nil, err
	}
	type pendingAgg struct {
		amount float64
		count  int
	}
	bucket := map[string]*pendingAgg{}
	for _, p := range pending {
		b, ok := bucket[p.Currency]
		if !ok {
			b = &pendingAgg{}
			bucket[p.Currency] = b
		}
		b.amount += p.Amount
		b.count++
	}
	for currency, b := range bucket {
		found := false
		for i := range stats {
			if stats[i].Currency == currency {
				stats[i].PendingAmount = b.amount
				stats[i].PendingCount = b.count
				found = true
				break
			}
		}
		if !found {
			stats = append(stats, model.CurrencyStats{
				Currency:      currency,
				PendingAmount: b.amount,
				PendingCount:  b.count,
				TopBusinesses: []model.TopBusiness{},
			})
		}
	}
	sort.Slice(stats, func(i, j int) bool {
		ki := stats[i].Total + stats[i].PendingAmount
		kj := stats[j].Total + stats[j].PendingAmount
		if ki != kj {
			return ki > kj
		}
		return stats[i].Currency < stats[j].Currency
	})
	return map[string]any{"stats": stats}, nil
}

func (h *handlers) createInvoice(ctx context.Context, params map[string]any) (any, error) {
	var inv model.Invoice
	if err := decodeInto(params, "invoice", &inv); err != nil {
		return nil, err
	}
	if inv.BusinessID != "" {
		if _, ok, err := h.businesses.Get(ctx, inv.BusinessID); err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("business not found")
		}
	}
	if inv.IssuedAt <= 0 {
		return nil, errors.New("issuedAt is required")
	}
	if inv.DueAt <= 0 {
		return nil, errors.New("dueAt is required")
	}
	if inv.Currency == "" {
		inv.Currency = defaultCurrency
	}
	now := nowUnix()
	inv.ID = newID()
	inv.CreatedAt = now
	inv.UpdatedAt = now
	if err := h.invoices.Insert(ctx, inv); err != nil {
		return nil, err
	}
	validated, err := h.validateUserIDs(ctx, inv.UserIDs)
	if err != nil {
		return nil, err
	}
	if err := h.userLinks.ReplaceForInvoice(ctx, inv.ID, validated); err != nil {
		return nil, err
	}
	inv.UserIDs = validated
	validatedTags, err := h.validateTagIDs(ctx, inv.TagIDs)
	if err != nil {
		return nil, err
	}
	if err := h.invoiceTags.ReplaceForInvoice(ctx, inv.ID, validatedTags); err != nil {
		return nil, err
	}
	inv.TagIDs = validatedTags
	return inv, nil
}

func (h *handlers) updateInvoice(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	var inv model.Invoice
	if err := decodeInto(params, "invoice", &inv); err != nil {
		return nil, err
	}
	if inv.BusinessID != "" {
		if _, ok, err := h.businesses.Get(ctx, inv.BusinessID); err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("business not found")
		}
	}
	if inv.IssuedAt <= 0 {
		return nil, errors.New("issuedAt is required")
	}
	if inv.DueAt <= 0 {
		return nil, errors.New("dueAt is required")
	}
	if inv.Currency == "" {
		inv.Currency = defaultCurrency
	}
	existing, ok, err := h.invoices.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("invoice not found")
	}
	inv.ID = id
	inv.CreatedAt = existing.CreatedAt
	inv.UpdatedAt = nowUnix()
	if _, err := h.invoices.Update(ctx, inv); err != nil {
		return nil, err
	}
	validated, err := h.validateUserIDs(ctx, inv.UserIDs)
	if err != nil {
		return nil, err
	}
	if err := h.userLinks.ReplaceForInvoice(ctx, inv.ID, validated); err != nil {
		return nil, err
	}
	inv.UserIDs = validated
	validatedTags, err := h.validateTagIDs(ctx, inv.TagIDs)
	if err != nil {
		return nil, err
	}
	if err := h.invoiceTags.ReplaceForInvoice(ctx, inv.ID, validatedTags); err != nil {
		return nil, err
	}
	inv.TagIDs = validatedTags
	return inv, nil
}

func (h *handlers) validateUserIDs(ctx context.Context, requested []string) ([]string, error) {
	if len(requested) == 0 {
		return []string{}, nil
	}
	users, err := h.users.List(ctx)
	if err != nil {
		return nil, err
	}
	valid := make(map[string]struct{}, len(users))
	for _, u := range users {
		valid[u.ID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(requested))
	out := make([]string, 0, len(requested))
	for _, id := range requested {
		if _, ok := valid[id]; !ok {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func (h *handlers) validateTagIDs(ctx context.Context, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}
	tags, err := h.tags.List(ctx)
	if err != nil {
		return nil, err
	}
	valid := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		valid[t.ID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := valid[id]; !ok {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func (h *handlers) deleteInvoice(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)

	upcomingIDs, err := h.upcomingLinks.UpcomingIDsByInvoice(ctx, id)
	if err != nil {
		return nil, err
	}
	// Delete the upcoming parents first so their link rows die via FK CASCADE;
	// a failure here leaves the invoice intact and the user can safely retry.
	for _, upID := range upcomingIDs {
		if err := h.upcoming.Delete(ctx, upID); err != nil {
			return nil, err
		}
	}
	if err := h.invoices.Delete(ctx, id); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

// ── RecurringInvoice ──────────────────────────────────────────────────────────

func (h *handlers) listRecurring(ctx context.Context) (any, error) {
	if err := h.reconciler.Catchup(ctx); err != nil {
		return nil, err
	}
	items, err := h.recurring.List(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	mapping, err := h.recurringUserLinks.UserIDsBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].UserIDs = mapping[items[i].ID]
	}
	tagMapping, err := h.recurringTags.TagIDsBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].TagIDs = tagMapping[items[i].ID]
	}
	return items, nil
}

func (h *handlers) createRecurring(ctx context.Context, params map[string]any) (any, error) {
	var r model.RecurringInvoice
	if err := decodeInto(params, "recurring", &r); err != nil {
		return nil, err
	}
	if r.BusinessID != "" {
		if _, ok, err := h.businesses.Get(ctx, r.BusinessID); err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("business not found")
		}
	}
	if !validFrequency(r.Frequency) {
		return nil, errors.New("frequency must be one of daily, weekly, monthly, yearly")
	}
	if err := validateAnchors(r); err != nil {
		return nil, err
	}
	if r.StartAt <= 0 {
		return nil, errors.New("startAt is required")
	}
	if r.Currency == "" {
		r.Currency = defaultCurrency
	}
	now := nowUnix()
	r.ID = newID()
	r.CreatedAt = now
	r.UpdatedAt = now
	if err := h.recurring.Insert(ctx, r); err != nil {
		return nil, err
	}
	validated, err := h.validateUserIDs(ctx, r.UserIDs)
	if err != nil {
		return nil, err
	}
	if err := h.recurringUserLinks.ReplaceForRecurring(ctx, r.ID, validated); err != nil {
		return nil, err
	}
	r.UserIDs = validated
	validatedTags, err := h.validateTagIDs(ctx, r.TagIDs)
	if err != nil {
		return nil, err
	}
	if err := h.recurringTags.ReplaceForRecurring(ctx, r.ID, validatedTags); err != nil {
		return nil, err
	}
	r.TagIDs = validatedTags
	if err := h.reconciler.Reconcile(ctx, r.ID); err != nil {
		return nil, err
	}
	return r, nil
}

func (h *handlers) updateRecurring(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	var r model.RecurringInvoice
	if err := decodeInto(params, "recurring", &r); err != nil {
		return nil, err
	}
	if r.BusinessID != "" {
		if _, ok, err := h.businesses.Get(ctx, r.BusinessID); err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("business not found")
		}
	}
	if !validFrequency(r.Frequency) {
		return nil, errors.New("frequency must be one of daily, weekly, monthly, yearly")
	}
	if err := validateAnchors(r); err != nil {
		return nil, err
	}
	if r.StartAt <= 0 {
		return nil, errors.New("startAt is required")
	}
	if r.Currency == "" {
		r.Currency = defaultCurrency
	}
	existing, ok, err := h.recurring.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("recurring not found")
	}
	r.ID = id
	r.CreatedAt = existing.CreatedAt
	r.UpdatedAt = nowUnix()
	if _, err := h.recurring.Update(ctx, r); err != nil {
		return nil, err
	}
	validated, err := h.validateUserIDs(ctx, r.UserIDs)
	if err != nil {
		return nil, err
	}
	// Rewrite the link rows before Reconcile so the wipe+re-emit picks up the new user list.
	if err := h.recurringUserLinks.ReplaceForRecurring(ctx, id, validated); err != nil {
		return nil, err
	}
	r.UserIDs = validated
	validatedTags, err := h.validateTagIDs(ctx, r.TagIDs)
	if err != nil {
		return nil, err
	}
	// Rewrite the tag links before Reconcile so the wipe+re-emit picks up the new tag list.
	if err := h.recurringTags.ReplaceForRecurring(ctx, id, validatedTags); err != nil {
		return nil, err
	}
	r.TagIDs = validatedTags
	if err := h.reconciler.Reconcile(ctx, id); err != nil {
		return nil, err
	}
	return r, nil
}

func (h *handlers) deleteRecurring(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	if err := h.reconciler.WipeOnly(ctx, id); err != nil {
		return nil, err
	}
	if err := h.recurring.Delete(ctx, id); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

// ── UpcomingInvoice ───────────────────────────────────────────────────────────

func (h *handlers) listUpcoming(ctx context.Context) (any, error) {
	if err := h.reconciler.Catchup(ctx); err != nil {
		return nil, err
	}
	items, err := h.upcoming.List(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	mapping, err := h.upcomingUserLinks.UserIDsBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].UserIDs = mapping[items[i].ID]
	}
	tagMapping, err := h.upcomingTags.TagIDsBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].TagIDs = tagMapping[items[i].ID]
	}
	return items, nil
}

func (h *handlers) createUpcoming(ctx context.Context, params map[string]any) (any, error) {
	var u model.UpcomingInvoice
	if err := decodeInto(params, "upcoming", &u); err != nil {
		return nil, err
	}
	if u.BusinessID != "" {
		if _, ok, err := h.businesses.Get(ctx, u.BusinessID); err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("business not found")
		}
	}
	if u.DueAt <= 0 {
		return nil, errors.New("dueAt is required")
	}
	if u.Currency == "" {
		u.Currency = defaultCurrency
	}
	now := nowUnix()
	u.ID = newID()
	u.CreatedAt = now
	u.UpdatedAt = now
	if err := h.upcoming.Insert(ctx, u); err != nil {
		return nil, err
	}
	validated, err := h.validateUserIDs(ctx, u.UserIDs)
	if err != nil {
		return nil, err
	}
	if err := h.upcomingUserLinks.ReplaceForUpcoming(ctx, u.ID, validated); err != nil {
		return nil, err
	}
	u.UserIDs = validated
	validatedTags, err := h.validateTagIDs(ctx, u.TagIDs)
	if err != nil {
		return nil, err
	}
	if err := h.upcomingTags.ReplaceForUpcoming(ctx, u.ID, validatedTags); err != nil {
		return nil, err
	}
	u.TagIDs = validatedTags
	if err := h.reconciler.ReconcileUpcoming(ctx, u.ID); err != nil {
		return nil, err
	}
	return u, nil
}

func (h *handlers) updateUpcoming(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	var u model.UpcomingInvoice
	if err := decodeInto(params, "upcoming", &u); err != nil {
		return nil, err
	}
	if u.BusinessID != "" {
		if _, ok, err := h.businesses.Get(ctx, u.BusinessID); err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("business not found")
		}
	}
	if u.DueAt <= 0 {
		return nil, errors.New("dueAt is required")
	}
	if u.Currency == "" {
		u.Currency = defaultCurrency
	}
	existing, ok, err := h.upcoming.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("upcoming not found")
	}
	u.ID = id
	u.CreatedAt = existing.CreatedAt
	u.UpdatedAt = nowUnix()
	if _, err := h.upcoming.Update(ctx, u); err != nil {
		return nil, err
	}
	validated, err := h.validateUserIDs(ctx, u.UserIDs)
	if err != nil {
		return nil, err
	}
	// Rewrite the link rows before Reconcile so the wipe+re-emit picks up the new user list.
	if err := h.upcomingUserLinks.ReplaceForUpcoming(ctx, id, validated); err != nil {
		return nil, err
	}
	u.UserIDs = validated
	validatedTags, err := h.validateTagIDs(ctx, u.TagIDs)
	if err != nil {
		return nil, err
	}
	// Rewrite the tag links before Reconcile so the wipe+re-emit picks up the new tag list.
	if err := h.upcomingTags.ReplaceForUpcoming(ctx, id, validatedTags); err != nil {
		return nil, err
	}
	u.TagIDs = validatedTags
	if err := h.reconciler.ReconcileUpcoming(ctx, id); err != nil {
		return nil, err
	}
	return u, nil
}

func (h *handlers) deleteUpcoming(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	if err := h.reconciler.WipeOnlyUpcoming(ctx, id); err != nil {
		return nil, err
	}
	if err := h.upcoming.Delete(ctx, id); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

// ── User ──────────────────────────────────────────────────────────────────────

func (h *handlers) listUser(ctx context.Context) (any, error) {
	return h.users.List(ctx)
}

func (h *handlers) createUser(ctx context.Context, params map[string]any) (any, error) {
	var u model.User
	if err := decodeInto(params, "user", &u); err != nil {
		return nil, err
	}
	u.Name = strings.TrimSpace(u.Name)
	if u.Name == "" {
		return nil, errors.New("name is required")
	}
	now := nowUnix()
	u.ID = newID()
	u.CreatedAt = now
	u.UpdatedAt = now
	if err := h.users.Insert(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (h *handlers) updateUser(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	var u model.User
	if err := decodeInto(params, "user", &u); err != nil {
		return nil, err
	}
	u.Name = strings.TrimSpace(u.Name)
	if u.Name == "" {
		return nil, errors.New("name is required")
	}
	existing, ok, err := h.users.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("user not found")
	}
	u.ID = id
	u.CreatedAt = existing.CreatedAt
	u.UpdatedAt = nowUnix()
	if _, err := h.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (h *handlers) deleteUser(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	if err := h.users.Delete(ctx, id); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

// ── Tag ───────────────────────────────────────────────────────────────────────

func (h *handlers) listTag(ctx context.Context) (any, error) {
	return h.tags.List(ctx)
}

func (h *handlers) createTag(ctx context.Context, params map[string]any) (any, error) {
	var t model.Tag
	if err := decodeInto(params, "tag", &t); err != nil {
		return nil, err
	}
	t.Name = strings.TrimSpace(t.Name)
	if t.Name == "" {
		return nil, errors.New("name is required")
	}
	if _, ok, err := h.tags.GetByName(ctx, t.Name); err != nil {
		return nil, err
	} else if ok {
		return nil, errors.New("tag name already exists")
	}
	now := nowUnix()
	t.ID = newID()
	t.CreatedAt = now
	t.UpdatedAt = now
	if err := h.tags.Insert(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (h *handlers) updateTag(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	var t model.Tag
	if err := decodeInto(params, "tag", &t); err != nil {
		return nil, err
	}
	t.Name = strings.TrimSpace(t.Name)
	if t.Name == "" {
		return nil, errors.New("name is required")
	}
	existing, ok, err := h.tags.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("tag not found")
	}
	if found, ok, err := h.tags.GetByName(ctx, t.Name); err != nil {
		return nil, err
	} else if ok && found.ID != id {
		return nil, errors.New("tag name already exists")
	}
	t.ID = id
	t.CreatedAt = existing.CreatedAt
	t.UpdatedAt = nowUnix()
	if _, err := h.tags.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (h *handlers) deleteTag(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	if err := h.tags.Delete(ctx, id); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

// ── Upload Sessions & Attachments ─────────────────────────────────────────────

func (h *handlers) createUploadSession(ctx context.Context, params map[string]any) (any, error) {
	invoiceID, _ := params["invoiceId"].(string)
	ttl := 600.0
	if v, ok := params["ttlSeconds"].(float64); ok {
		ttl = v
	}
	if ttl < 60 {
		ttl = 60
	} else if ttl > 3600 {
		ttl = 3600
	}
	now := nowUnix()
	session := model.UploadSession{
		ID:        newID(),
		InvoiceID: invoiceID,
		Status:    "active",
		CreatedAt: now,
		ExpiresAt: now + int64(ttl),
	}
	if err := h.uploadSessions.Insert(ctx, session); err != nil {
		return nil, err
	}
	return map[string]any{"token": session.ID, "expiresAt": session.ExpiresAt}, nil
}

func (h *handlers) getUploadSession(ctx context.Context, params map[string]any) (any, error) {
	token, _ := params["token"].(string)
	session, ok, err := h.uploadSessions.Get(ctx, token)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("session not found")
	}
	attachments, err := h.attachments.ListBySession(ctx, session.ID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"session": session, "attachments": attachments}, nil
}

var allowedUploadMimes = map[string]struct{}{
	"image/png":       {},
	"image/jpeg":      {},
	"image/webp":      {},
	"application/pdf": {},
}

const maxAttachmentBytes = 15 * 1024 * 1024

func (h *handlers) mobileUploadFile(ctx context.Context, params map[string]any) (any, error) {
	token, _ := params["token"].(string)
	filename, _ := params["filename"].(string)
	contentType, _ := params["contentType"].(string)
	dataBase64, _ := params["dataBase64"].(string)

	session, ok, err := h.uploadSessions.Get(ctx, token)
	if err != nil {
		return nil, err
	}
	if !ok || session.Status != "active" || session.ExpiresAt <= nowUnix() {
		return nil, errors.New("session expired or not found")
	}

	if _, ok := allowedUploadMimes[contentType]; !ok {
		return nil, errors.New("unsupported file type")
	}

	decoded, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return nil, errors.New("invalid base64")
	}
	if len(decoded) > maxAttachmentBytes {
		return nil, errors.New("file too large (max 15 MB)")
	}

	sanitisedFilename := filepath.Base(filename)

	ext := extFromMime(contentType)

	var folder string
	if session.InvoiceID != "" {
		folder = filepath.Join(h.dataDir, "invoices", session.InvoiceID)
	} else {
		folder = filepath.Join(h.dataDir, "invoices", session.ID)
	}

	if err := os.MkdirAll(folder, 0o755); err != nil {
		return nil, err
	}

	attachmentID := newID()
	finalPath := filepath.Join(folder, attachmentID+"."+ext)
	tmpPath := finalPath + ".tmp"
	if err := os.WriteFile(tmpPath, decoded, 0o644); err != nil {
		return nil, err
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return nil, err
	}

	a := model.Attachment{
		ID:               attachmentID,
		InvoiceID:        session.InvoiceID,
		SessionID:        session.ID,
		Mime:             contentType,
		OriginalFilename: sanitisedFilename,
		SizeBytes:        int64(len(decoded)),
		CreatedAt:        nowUnix(),
	}
	if err := h.attachments.Insert(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (h *handlers) listInvoiceAttachments(ctx context.Context, params map[string]any) (any, error) {
	invoiceID, _ := params["invoiceId"].(string)
	if invoiceID == "" {
		return nil, errors.New("invoiceId is required")
	}
	return h.attachments.ListByInvoice(ctx, invoiceID)
}

func (h *handlers) attachSessionToInvoice(ctx context.Context, params map[string]any) (any, error) {
	token, _ := params["token"].(string)
	invoiceID, _ := params["invoiceId"].(string)
	if token == "" {
		return nil, errors.New("token is required")
	}
	if invoiceID == "" {
		return nil, errors.New("invoiceId is required")
	}

	session, ok, err := h.uploadSessions.Get(ctx, token)
	if err != nil {
		return nil, err
	}
	if !ok || session.Status != "active" {
		return nil, errors.New("session not found or not active")
	}

	currentFolder := filepath.Join(h.dataDir, "invoices", session.ID)
	targetFolder := filepath.Join(h.dataDir, "invoices", invoiceID)

	if currentFolder != targetFolder {
		if _, statErr := os.Stat(currentFolder); statErr == nil {
			if err := os.MkdirAll(targetFolder, 0o755); err != nil {
				return nil, err
			}
			entries, err := os.ReadDir(currentFolder)
			if err != nil {
				return nil, err
			}
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				src := filepath.Join(currentFolder, e.Name())
				dst := filepath.Join(targetFolder, e.Name())
				if err := os.Rename(src, dst); err != nil {
					return nil, err
				}
			}
		}
	}

	count, err := h.attachments.AttachToInvoice(ctx, session.ID, invoiceID)
	if err != nil {
		return nil, err
	}
	if err := h.uploadSessions.MarkConsumed(ctx, session.ID); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "count": count}, nil
}

func (h *handlers) getAttachmentBytes(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	if id == "" {
		return nil, errors.New("id is required")
	}
	a, ok, err := h.attachments.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("attachment not found")
	}
	var folder string
	if a.InvoiceID != "" {
		folder = filepath.Join(h.dataDir, "invoices", a.InvoiceID)
	} else {
		folder = filepath.Join(h.dataDir, "invoices", a.SessionID)
	}
	fullPath := filepath.Join(folder, a.ID+"."+extFromMime(a.Mime))
	b, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	if len(b) > 20*1024*1024 {
		return nil, errors.New("attachment too large to serve")
	}
	return map[string]any{
		"id":               a.ID,
		"mime":             a.Mime,
		"originalFilename": a.OriginalFilename,
		"dataBase64":       base64.StdEncoding.EncodeToString(b),
	}, nil
}

func (h *handlers) deleteAttachment(ctx context.Context, params map[string]any) (any, error) {
	id, _ := params["id"].(string)
	if id == "" {
		return nil, errors.New("id is required")
	}

	a, ok, err := h.attachments.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if ok {
		var folder string
		if a.InvoiceID != "" {
			folder = filepath.Join(h.dataDir, "invoices", a.InvoiceID)
		} else {
			folder = filepath.Join(h.dataDir, "invoices", a.SessionID)
		}
		ext := extFromMime(a.Mime)
		fullPath := filepath.Join(folder, a.ID+"."+ext)
		if rmErr := os.Remove(fullPath); rmErr != nil && !os.IsNotExist(rmErr) {
			return nil, rmErr
		}
	}

	if err := h.attachments.Delete(ctx, id); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}
