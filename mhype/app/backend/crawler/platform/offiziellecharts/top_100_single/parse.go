package top_100_single

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"mhype-plugin/crawler"
)

const baseURL = "https://www.offiziellecharts.de"

var (
	periodRE  = regexp.MustCompile(`\b(\d{2}\.\d{2}\.\d{4})\s*[-–]\s*(\d{2}\.\d{2}\.\d{4})\b`)
	coverRE   = regexp.MustCompile(`url\(['"]?([^'")]+)['"]?\)`)
	weeksRE   = regexp.MustCompile(`(\d+)`)
	wsRunRE   = regexp.MustCompile(`\s+`)
	xSplitRE  = regexp.MustCompile(`(?i)\s+x\s+`)
	featSplit = regexp.MustCompile(`(?i)\s+(?:feat\.?|ft\.?)\s+`)
)

func extract(body []byte, scrapedAtMs int64) (crawler.CrawlResult, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return crawler.CrawlResult{}, fmt.Errorf("parse html: %w", err)
	}
	start, end, err := findChartPeriod(doc)
	if err != nil {
		return crawler.CrawlResult{}, fmt.Errorf("chart period not found: %w", err)
	}
	bundleKey := fmt.Sprintf("%d_%d", start.Unix(), end.Unix())

	rows := findRows(doc)
	items := make([]crawler.Item, 0, len(rows))
	for _, r := range rows {
		rd, ok := extractRow(r)
		if !ok {
			continue
		}
		item, ok := buildItem(rd, start, end, scrapedAtMs)
		if !ok {
			continue
		}
		items = append(items, item)
	}
	return crawler.CrawlResult{Items: items, BundleKey: bundleKey}, nil
}

type rowData struct {
	position   int
	title      string
	artistRaw  string
	prevPos    int
	peakPos    int
	weeks      int
	sourceURL  string
	artworkURL string
}

func findChartPeriod(doc *html.Node) (time.Time, time.Time, error) {
	var start, end time.Time
	var found bool
	walkNodes(doc, func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}
		m := periodRE.FindStringSubmatch(textContent(n))
		if m == nil {
			return false
		}
		s, err1 := time.ParseInLocation("02.01.2006", m[1], time.UTC)
		e, err2 := time.ParseInLocation("02.01.2006", m[2], time.UTC)
		if err1 != nil || err2 != nil {
			return false
		}
		start, end, found = s, e, true
		return true
	})
	if !found {
		return time.Time{}, time.Time{}, fmt.Errorf("no dd.mm.yyyy date range in document")
	}
	return start, end, nil
}

func findRows(doc *html.Node) []*html.Node {
	var rows []*html.Node
	walkNodes(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "tr" && hasClass(n, "drill-down-link") {
			rows = append(rows, n)
		}
		return false
	})
	return rows
}

func extractRow(row *html.Node) (rowData, bool) {
	var rd rowData

	if posNode := firstMatching(row, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "span" && hasClass(n, "this-week")
	}); posNode != nil {
		p, err := strconv.Atoi(strings.TrimSpace(textContent(posNode)))
		if err != nil {
			return rowData{}, false
		}
		rd.position = p
	} else {
		return rowData{}, false
	}

	if titleNode := firstMatching(row, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "span" && hasClass(n, "info-title")
	}); titleNode != nil {
		rd.title = collapseWS(textContent(titleNode))
	}
	if rd.title == "" {
		return rowData{}, false
	}

	if artistNode := firstMatching(row, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "span" && hasClass(n, "info-artist")
	}); artistNode != nil {
		rd.artistRaw = collapseWS(textContent(artistNode))
	}
	if rd.artistRaw == "" {
		return rowData{}, false
	}

	if lwNode := firstMatching(row, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "span" && hasClass(n, "last-week")
	}); lwNode != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(textContent(lwNode))); err == nil {
			rd.prevPos = v
		}
	}

	walkNodes(row, func(n *html.Node) bool {
		if n.Type != html.ElementNode || n.Data != "span" || !hasClass(n, "plus-data") {
			return false
		}
		txt := collapseWS(textContent(n))
		switch {
		case strings.HasPrefix(txt, "In Charts:"):
			if m := weeksRE.FindString(txt); m != "" {
				if v, err := strconv.Atoi(m); err == nil {
					rd.weeks = v
				}
			}
		case strings.HasPrefix(txt, "Peak:"):
			if m := weeksRE.FindString(txt); m != "" {
				if v, err := strconv.Atoi(m); err == nil {
					rd.peakPos = v
				}
			}
		}
		return false
	})

	if a := firstMatching(row, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a" && hasClass(n, "drill-down")
	}); a != nil {
		if href := strings.TrimSpace(attr(a, "href")); href != "" {
			if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
				rd.sourceURL = href
			} else {
				rd.sourceURL = baseURL + href
			}
		}
	}

	if cover := firstMatching(row, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "span" && hasClass(n, "cover-img")
	}); cover != nil {
		if m := coverRE.FindStringSubmatch(attr(cover, "style")); m != nil {
			img := strings.TrimSpace(m[1])
			if strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") {
				rd.artworkURL = img
			} else if img != "" {
				rd.artworkURL = baseURL + img
			}
		}
	}

	return rd, true
}

func splitArtists(s string) []string {
	parts := []string{s}
	parts = splitMany(parts, func(p string) []string { return featSplit.Split(p, -1) })
	parts = splitMany(parts, func(p string) []string { return strings.Split(p, " & ") })
	parts = splitMany(parts, func(p string) []string { return xSplitRE.Split(p, -1) })
	parts = splitMany(parts, func(p string) []string { return strings.Split(p, ", ") })
	parts = splitMany(parts, func(p string) []string { return strings.Split(p, " / ") })

	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func splitMany(in []string, split func(string) []string) []string {
	out := make([]string, 0, len(in))
	for _, p := range in {
		out = append(out, split(p)...)
	}
	return out
}

func buildItem(r rowData, periodStart, periodEnd time.Time, scrapedAtMs int64) (crawler.Item, bool) {
	artists := splitArtists(r.artistRaw)
	if len(artists) == 0 || r.title == "" || r.position < 1 || r.position > 100 {
		return crawler.Item{}, false
	}

	sourceURL := r.sourceURL
	if sourceURL == "" {
		sourceURL = chartURL
	}

	item := crawler.Item{
		ID:         inlineItemID(crawlerID, periodStart.Unix(), r.position, r.title, artists[0]),
		Kind:       crawler.KindChartEntry,
		CrawlerID:  crawlerID,
		SourceURL:  sourceURL,
		Title:      r.title,
		ScrapedAt:  scrapedAtMs,
		Artists:    artists,
		ArtworkURL: r.artworkURL,
		Chart: &crawler.ChartContext{
			Name:         chartName,
			Position:     r.position,
			PrevPosition: r.prevPos,
			PeakPosition: r.peakPos,
			WeeksOnChart: r.weeks,
			PeriodStart:  periodStart.Format("2006-01-02"),
			PeriodEnd:    periodEnd.Format("2006-01-02"),
		},
	}
	return item, true
}

func inlineItemID(crawlerID string, startUnix int64, position int, title, primaryArtist string) string {
	key := strings.Join([]string{
		crawlerID,
		strconv.FormatInt(startUnix, 10),
		strconv.Itoa(position),
		title,
		primaryArtist,
	}, "|")
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func walkNodes(n *html.Node, fn func(*html.Node) bool) {
	if fn(n) {
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkNodes(c, fn)
	}
}

func textContent(n *html.Node) string {
	if n == nil {
		return ""
	}
	var sb strings.Builder
	walkNodes(n, func(node *html.Node) bool {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
		}
		return false
	})
	return strings.TrimSpace(sb.String())
}

func firstMatching(n *html.Node, pred func(*html.Node) bool) *html.Node {
	var found *html.Node
	walkNodes(n, func(node *html.Node) bool {
		if pred(node) {
			found = node
			return true
		}
		return false
	})
	return found
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, want string) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, c := range strings.Fields(attr(n, "class")) {
		if c == want {
			return true
		}
	}
	return false
}

func collapseWS(s string) string {
	return strings.TrimSpace(wsRunRE.ReplaceAllString(s, " "))
}
