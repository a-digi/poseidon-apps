package hot_100

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"mhype-plugin/crawler"
)

var (
	wsRunRE  = regexp.MustCompile(`\s+`)
	xSplitRE = regexp.MustCompile(`(?i)\s+x\s+`)
	featSplit = regexp.MustCompile(`(?i)\s+(?:feat\.?|ft\.?)\s+`)
	digitRE  = regexp.MustCompile(`^\d+$`)
	isoDateRE = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`)
)

func extract(body []byte, scrapedAtMs int64) (crawler.CrawlResult, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return crawler.CrawlResult{}, err
	}

	period := findChartPeriod(doc)

	rows := findRows(doc)
	items := make([]crawler.Item, 0, len(rows))
	for _, row := range rows {
		item, ok := extractRow(row, period, scrapedAtMs)
		if !ok {
			continue
		}
		items = append(items, item)
	}
	return crawler.CrawlResult{Items: items, BundleKey: "current", Snapshot: true}, nil
}

// findRows returns all <ul> nodes whose class contains "o-chart-results-list-row".
func findRows(doc *html.Node) []*html.Node {
	var rows []*html.Node
	walkNodes(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "ul" && hasClass(n, "o-chart-results-list-row") {
			rows = append(rows, n)
		}
		return false
	})
	return rows
}

// findChartPeriod scans the document for "Week of" text or a <time datetime="..."> element.
// Returns "2006-01-02" formatted string, or empty if not found.
func findChartPeriod(doc *html.Node) string {
	var period string
	walkNodes(doc, func(n *html.Node) bool {
		if period != "" {
			return true
		}
		// Check <time datetime="...">
		if n.Type == html.ElementNode && n.Data == "time" {
			if dt := attr(n, "datetime"); isoDateRE.MatchString(dt) {
				if m := isoDateRE.FindString(dt); m != "" {
					period = m
					return true
				}
			}
		}
		// Check text nodes containing "Week of"
		if n.Type == html.TextNode && strings.Contains(n.Data, "Week of") {
			if m := isoDateRE.FindString(n.Data); m != "" {
				period = m
				return true
			}
		}
		return false
	})
	// Also try element text content for "Week of ..." patterns
	if period == "" {
		walkNodes(doc, func(n *html.Node) bool {
			if period != "" {
				return true
			}
			if n.Type != html.ElementNode {
				return false
			}
			txt := textContent(n)
			if strings.Contains(txt, "Week of") {
				// Try to parse a date like "Week of May 10, 2025" by looking for an ISO date in nearby attrs
				// or look for a sibling <time> element
				if m := isoDateRE.FindString(txt); m != "" {
					period = m
					return true
				}
				// Try to find a <time> child
				if t := firstMatching(n, func(c *html.Node) bool {
					return c.Type == html.ElementNode && c.Data == "time" && isoDateRE.MatchString(attr(c, "datetime"))
				}); t != nil {
					if m := isoDateRE.FindString(attr(t, "datetime")); m != "" {
						period = m
						return true
					}
				}
			}
			return false
		})
	}
	return period
}

type rowData struct {
	position   int
	title      string
	artistRaw  string
	prevPos    int
	peakPos    int
	weeks      int
	artworkURL string
}

func extractRow(row *html.Node, period string, scrapedAtMs int64) (crawler.Item, bool) {
	rd := parseRowData(row)
	if rd.position < 1 || rd.position > 100 || rd.title == "" || rd.artistRaw == "" {
		return crawler.Item{}, false
	}

	artists := splitArtists(rd.artistRaw)
	if len(artists) == 0 {
		return crawler.Item{}, false
	}

	chart := &crawler.ChartContext{
		Name:         chartName,
		Position:     rd.position,
		PrevPosition: rd.prevPos,
		PeakPosition: rd.peakPos,
		WeeksOnChart: rd.weeks,
	}
	if period != "" {
		chart.PeriodStart = period
		chart.PeriodEnd = period
	}

	item := crawler.Item{
		ID:         inlineItemID(crawlerID, dayBucket(scrapedAtMs), rd.position, rd.title, artists[0]),
		Kind:       crawler.KindChartEntry,
		CrawlerID:  crawlerID,
		SourceURL:  chartURL,
		Title:      rd.title,
		ScrapedAt:  scrapedAtMs,
		Artists:    artists,
		ArtworkURL: rd.artworkURL,
		Chart:      chart,
	}
	return item, true
}

func parseRowData(row *html.Node) rowData {
	var rd rowData

	// Position: first <li> child → first <span> with class "c-label a-font-c-label".
	// Fall back to any span with just "c-label" that contains a pure digit.
	if posSpan := firstMatching(row, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "span" &&
			hasClass(n, "c-label") && hasClass(n, "a-font-c-label")
	}); posSpan != nil {
		if v, err := strconv.Atoi(strings.TrimSpace(textContent(posSpan))); err == nil {
			rd.position = v
		}
	}
	if rd.position == 0 {
		// Fallback: first span with "c-label" whose text is a pure digit
		firstMatching(row, func(n *html.Node) bool {
			if n.Type != html.ElementNode || n.Data != "span" || !hasClass(n, "c-label") {
				return false
			}
			txt := strings.TrimSpace(textContent(n))
			if digitRE.MatchString(txt) {
				if v, err := strconv.Atoi(txt); err == nil {
					rd.position = v
					return true
				}
			}
			return false
		})
	}

	// Title: <h3 id="title-of-a-story"> OR <h3 class="...c-title...">
	if titleNode := firstMatching(row, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "h3" && attr(n, "id") == "title-of-a-story"
	}); titleNode != nil {
		rd.title = collapseWS(textContent(titleNode))
	}
	if rd.title == "" {
		if titleNode := firstMatching(row, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "h3" && hasClass(n, "c-title")
		}); titleNode != nil {
			rd.title = collapseWS(textContent(titleNode))
		}
	}

	// Artist: <span class="...c-label...a-no-trucate..."> OR <p class="...c-tagline...">
	if artistNode := firstMatching(row, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "span" &&
			hasClass(n, "c-label") && hasClass(n, "a-no-trucate")
	}); artistNode != nil {
		rd.artistRaw = collapseWS(textContent(artistNode))
	}
	if rd.artistRaw == "" {
		if artistNode := firstMatching(row, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "p" && hasClass(n, "c-tagline")
		}); artistNode != nil {
			rd.artistRaw = collapseWS(textContent(artistNode))
		}
	}

	if imgNode := firstMatching(row, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "img"
	}); imgNode != nil {
		for _, key := range []string{"data-lazy-src", "src"} {
			u := strings.TrimSpace(attr(imgNode, key))
			if u == "" || strings.Contains(u, "lazyload-fallback") {
				continue
			}
			rd.artworkURL = u
			break
		}
	}

	// Chart stats from <li> children: each stat <li> contains a label span and a value span.
	// Walk direct <li> children of the row <ul>.
	for li := row.FirstChild; li != nil; li = li.NextSibling {
		if li.Type != html.ElementNode || li.Data != "li" {
			continue
		}
		var label, value string
		// Collect all spans in this <li>.
		var spans []*html.Node
		walkNodes(li, func(n *html.Node) bool {
			if n.Type == html.ElementNode && n.Data == "span" {
				spans = append(spans, n)
			}
			return false
		})
		// Expect at least two spans: first is label, second is value.
		for i, sp := range spans {
			txt := strings.TrimSpace(textContent(sp))
			switch txt {
			case "LAST WEEK", "PEAK POS.", "WKS ON CHART":
				label = txt
				// Value is the next span.
				if i+1 < len(spans) {
					value = strings.TrimSpace(textContent(spans[i+1]))
				}
			}
		}
		if label == "" || value == "" {
			continue
		}
		v, err := strconv.Atoi(value)
		if err != nil {
			continue
		}
		switch label {
		case "LAST WEEK":
			rd.prevPos = v
		case "PEAK POS.":
			rd.peakPos = v
		case "WKS ON CHART":
			rd.weeks = v
		}
	}

	return rd
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
