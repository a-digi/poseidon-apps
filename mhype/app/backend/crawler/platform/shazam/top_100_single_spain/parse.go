package top_100_single_spain

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"mhype-plugin/crawler"
)

var (
	wsRunRE   = regexp.MustCompile(`\s+`)
	featSplit = regexp.MustCompile(`(?i)\s+(?:feat\.?|ft\.?)\s+`)
	datePart  = regexp.MustCompile(`(?i)(\w+,\s+\d{1,2}\s+\w+\s+\d{4})`)
)

type csvRow struct {
	rank   int
	artist string
	title  string
}

func extract(csvBody, htmlBody []byte, scrapedAtMs int64) (crawler.CrawlResult, error) {
	rows, periodStart, periodEnd, err := parseCSV(csvBody)
	if err != nil {
		return crawler.CrawlResult{}, err
	}

	artworkByRank := parseArtwork(htmlBody)

	items := make([]crawler.Item, 0, limit)
	for _, row := range rows {
		if row.rank < 1 || row.title == "" || row.artist == "" {
			continue
		}
		if row.rank > limit {
			break
		}
		artists := splitArtists(row.artist)
		if len(artists) == 0 {
			continue
		}
		item := crawler.Item{
			ID:         inlineItemID(crawlerID, dayBucket(scrapedAtMs), row.rank, row.title, artists[0]),
			Kind:       crawler.KindChartEntry,
			CrawlerID:  crawlerID,
			SourceURL:  htmlURL,
			Title:      row.title,
			ScrapedAt:  scrapedAtMs,
			Artists:    artists,
			ArtworkURL: artworkByRank[row.rank],
			Chart: &crawler.ChartContext{
				Name:        chartName,
				Position:    row.rank,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
			},
		}
		items = append(items, item)
	}
	return crawler.CrawlResult{Items: items, BundleKey: "current", Snapshot: true}, nil
}

// parseCSV strips the UTF-8 BOM, skips the date header and the column header row,
// then returns all data rows up to limit, plus the chart period derived from the date header.
func parseCSV(body []byte) ([]csvRow, string, string, error) {
	body = bytes.TrimPrefix(body, []byte{0xEF, 0xBB, 0xBF})

	r := csv.NewReader(bytes.NewReader(body))
	r.FieldsPerRecord = -1 // allow variable field counts

	var (
		rows        []csvRow
		periodStart string
		periodEnd   string
		headerSeen  bool
		dateRaw     string
	)

	for {
		record, err := r.Read()
		if err != nil {
			break
		}
		if len(record) == 1 && dateRaw == "" {
			dateRaw = record[0]
			continue
		}
		if !headerSeen {
			if len(record) == 3 && record[0] == "Rank" {
				headerSeen = true
				periodStart, periodEnd = parsePeriod(dateRaw)
			}
			continue
		}
		if len(record) < 3 {
			continue
		}
		rank, err := strconv.Atoi(strings.TrimSpace(record[0]))
		if err != nil {
			continue
		}
		rows = append(rows, csvRow{
			rank:   rank,
			artist: strings.TrimSpace(record[1]),
			title:  strings.TrimSpace(record[2]),
		})
	}
	return rows, periodStart, periodEnd, nil
}

func parsePeriod(dateRaw string) (string, string) {
	m := datePart.FindString(dateRaw)
	if m == "" {
		return "", ""
	}
	t, err := time.Parse("Monday, 2 January 2006", m)
	if err != nil {
		return "", ""
	}
	end := t.Format("2006-01-02")
	start := t.Add(-7 * 24 * time.Hour).Format("2006-01-02")
	return start, end
}

// parseArtwork walks the HTML looking for <li> elements whose class contains
// "page_songItem__lAdHy". For each, it extracts the rank from the ranking-number
// span and the largest artwork URL from the img srcSet.
func parseArtwork(body []byte) map[int]string {
	if len(body) == 0 {
		return nil
	}
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil
	}

	artworkByRank := make(map[int]string)
	walkNodes(doc, func(n *html.Node) bool {
		if n.Type != html.ElementNode || n.Data != "div" {
			return false
		}
		if !hasClassContaining(n, "page_songItem__") {
			return false
		}
		rankSpan := firstMatching(n, func(c *html.Node) bool {
			return c.Type == html.ElementNode && c.Data == "span" &&
				hasClassContaining(c, "SongItem-module_rankingNumber")
		})
		if rankSpan == nil {
			return false
		}
		rank, err := strconv.Atoi(strings.TrimSpace(textContent(rankSpan)))
		if err != nil || rank < 1 {
			return false
		}
		imgNode := firstMatching(n, func(c *html.Node) bool {
			return c.Type == html.ElementNode && c.Data == "img" &&
				strings.Contains(attr(c, "srcset"), "mzstatic.com")
		})
		if imgNode == nil {
			return false
		}
		u := largestSrcSetURL(attr(imgNode, "srcset"))
		if u != "" {
			artworkByRank[rank] = u
		}
		return false
	})
	return artworkByRank
}

func largestSrcSetURL(srcset string) string {
	best := ""
	bestW := -1
	for _, part := range strings.Split(srcset, ",") {
		part = strings.TrimSpace(part)
		fields := strings.Fields(part)
		if len(fields) < 2 {
			continue
		}
		wStr := strings.TrimSuffix(fields[len(fields)-1], "w")
		w, err := strconv.Atoi(wStr)
		if err != nil || w <= bestW {
			continue
		}
		best = fields[0]
		bestW = w
	}
	return best
}

func dayBucket(ms int64) int64 { return ms / 86400000 }

func inlineItemID(crawlerID string, bucket int64, position int, title, primaryArtist string) string {
	key := strings.Join([]string{
		crawlerID,
		strconv.FormatInt(bucket, 10),
		strconv.Itoa(position),
		title,
		primaryArtist,
	}, "|")
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func splitArtists(s string) []string {
	parts := []string{s}
	parts = splitMany(parts, func(p string) []string { return featSplit.Split(p, -1) })
	parts = splitMany(parts, func(p string) []string { return strings.Split(p, " & ") })
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
		if found != nil {
			return true
		}
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

func hasClassContaining(n *html.Node, substr string) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, c := range strings.Fields(attr(n, "class")) {
		if strings.Contains(c, substr) {
			return true
		}
	}
	return false
}

func collapseWS(s string) string {
	return strings.TrimSpace(wsRunRE.ReplaceAllString(s, " "))
}
