package hot_100

import (
	"testing"
	"time"

	"mhype-plugin/crawler"
)

// fixture mirrors the actual Billboard Hot 100 HTML structure with two chart rows.
const fixture = `<!DOCTYPE html>
<html>
<head><title>Billboard Hot 100</title></head>
<body>
<div class="chart-results-list">

  <!-- Row 1 -->
  <ul class="o-chart-results-list-row">
    <li class="o-chart-results-list__item">
      <span class="c-label a-font-c-label a-font-c-label--medium">1</span>
    </li>
    <li class="o-chart-results-list__item lrv-u-flex">
      <div class="lrv-u-flex-shrink-0">
        <img src="https://charts-static.billboard.com/img/2024/01/sample-artist-001-ri7-180x180.jpg" alt="artwork" />
      </div>
      <div class="lrv-u-flex-direction-column">
        <h3 id="title-of-a-story" class="c-title a-no-trucate">Big Hit Song</h3>
        <span class="c-label a-no-trucate a-font-c-label--secondary">Sample Artist</span>
      </div>
    </li>
    <li class="o-chart-results-list__item lrv-u-text-align-center">
      <span class="c-label a-font-c-label a-font-c-label--medium">LAST WEEK</span>
      <span class="c-label a-font-c-label a-font-c-label--medium">2</span>
    </li>
    <li class="o-chart-results-list__item lrv-u-text-align-center">
      <span class="c-label a-font-c-label a-font-c-label--medium">PEAK POS.</span>
      <span class="c-label a-font-c-label a-font-c-label--medium">1</span>
    </li>
    <li class="o-chart-results-list__item lrv-u-text-align-center">
      <span class="c-label a-font-c-label a-font-c-label--medium">WKS ON CHART</span>
      <span class="c-label a-font-c-label a-font-c-label--medium">10</span>
    </li>
  </ul>

  <!-- Row 2 -->
  <ul class="o-chart-results-list-row">
    <li class="o-chart-results-list__item">
      <span class="c-label a-font-c-label a-font-c-label--medium">2</span>
    </li>
    <li class="o-chart-results-list__item lrv-u-flex">
      <div class="lrv-u-flex-direction-column">
        <h3 id="title-of-a-story" class="c-title a-no-trucate">Another Hit</h3>
        <span class="c-label a-no-trucate a-font-c-label--secondary">Second Artist feat. Guest</span>
      </div>
    </li>
    <li class="o-chart-results-list__item lrv-u-text-align-center">
      <span class="c-label a-font-c-label a-font-c-label--medium">LAST WEEK</span>
      <span class="c-label a-font-c-label a-font-c-label--medium">1</span>
    </li>
    <li class="o-chart-results-list__item lrv-u-text-align-center">
      <span class="c-label a-font-c-label a-font-c-label--medium">PEAK POS.</span>
      <span class="c-label a-font-c-label a-font-c-label--medium">1</span>
    </li>
    <li class="o-chart-results-list__item lrv-u-text-align-center">
      <span class="c-label a-font-c-label a-font-c-label--medium">WKS ON CHART</span>
      <span class="c-label a-font-c-label a-font-c-label--medium">5</span>
    </li>
  </ul>

</div>
</body>
</html>`

func TestExtract(t *testing.T) {
	scrapedAt := time.Now().UnixMilli()
	result, err := extract([]byte(fixture), scrapedAt)
	if err != nil {
		t.Fatalf("extract returned error: %v", err)
	}
	if !result.Snapshot {
		t.Fatal("Snapshot must be true")
	}
	if result.BundleKey != "current" {
		t.Fatalf("BundleKey: want %q got %q", "current", result.BundleKey)
	}
	if len(result.Items) < 1 {
		t.Fatal("expected at least 1 item")
	}

	it := result.Items[0]
	if it.Kind != crawler.KindChartEntry {
		t.Fatalf("Kind: want %q got %q", crawler.KindChartEntry, it.Kind)
	}
	if it.Title == "" {
		t.Fatal("Title must not be empty")
	}
	if len(it.Artists) == 0 {
		t.Fatal("Artists must not be empty")
	}
	if it.Chart == nil {
		t.Fatal("Chart context must not be nil")
	}
	if it.Chart.Position < 1 {
		t.Fatalf("Chart.Position: want >= 1 got %d", it.Chart.Position)
	}
}

func TestExtractTwoRows(t *testing.T) {
	result, err := extract([]byte(fixture), time.Now().UnixMilli())
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("want 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Chart.Position != 1 {
		t.Fatalf("first item position: want 1 got %d", result.Items[0].Chart.Position)
	}
	if result.Items[1].Chart.Position != 2 {
		t.Fatalf("second item position: want 2 got %d", result.Items[1].Chart.Position)
	}
	if result.Items[0].Title != "Big Hit Song" {
		t.Fatalf("first item title: want %q got %q", "Big Hit Song", result.Items[0].Title)
	}
	if result.Items[1].Title != "Another Hit" {
		t.Fatalf("second item title: want %q got %q", "Another Hit", result.Items[1].Title)
	}
}

func TestExtractArtistSplit(t *testing.T) {
	result, err := extract([]byte(fixture), time.Now().UnixMilli())
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(result.Items) < 2 {
		t.Fatal("need at least 2 items")
	}
	// "Second Artist feat. Guest" → ["Second Artist", "Guest"]
	second := result.Items[1]
	if len(second.Artists) != 2 {
		t.Fatalf("second item artists: want 2 got %v", second.Artists)
	}
	if second.Artists[0] != "Second Artist" {
		t.Fatalf("second.Artists[0]: want %q got %q", "Second Artist", second.Artists[0])
	}
	if second.Artists[1] != "Guest" {
		t.Fatalf("second.Artists[1]: want %q got %q", "Guest", second.Artists[1])
	}
}

func TestExtractChartStats(t *testing.T) {
	result, err := extract([]byte(fixture), time.Now().UnixMilli())
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(result.Items) < 1 {
		t.Fatal("need at least 1 item")
	}
	ch := result.Items[0].Chart
	if ch.PrevPosition != 2 {
		t.Fatalf("PrevPosition: want 2 got %d", ch.PrevPosition)
	}
	if ch.PeakPosition != 1 {
		t.Fatalf("PeakPosition: want 1 got %d", ch.PeakPosition)
	}
	if ch.WeeksOnChart != 10 {
		t.Fatalf("WeeksOnChart: want 10 got %d", ch.WeeksOnChart)
	}
}

func TestExtractSnapshot(t *testing.T) {
	result, _ := extract([]byte(fixture), time.Now().UnixMilli())
	if !result.Snapshot {
		t.Fatal("Snapshot must be true")
	}
	if result.BundleKey != "current" {
		t.Fatal("BundleKey must be 'current'")
	}
}

func TestExtractEmptyHTML(t *testing.T) {
	result, err := extract([]byte("<html><body></body></html>"), time.Now().UnixMilli())
	if err != nil {
		t.Fatalf("extract on empty page: %v", err)
	}
	if len(result.Items) != 0 {
		t.Fatalf("want 0 items, got %d", len(result.Items))
	}
}

func rowHTML(imgAttrs string) string {
	return `<html><body><ul class="o-chart-results-list-row">` +
		`<li><span class="c-label a-font-c-label">1</span></li>` +
		`<li><h3 id="title-of-a-story">Test Song</h3>` +
		`<span class="c-label a-no-trucate">Test Artist</span>` +
		`<img ` + imgAttrs + `></li>` +
		`</ul></body></html>`
}

func TestExtractArtwork(t *testing.T) {
	cases := []struct {
		name     string
		imgAttrs string
		wantURL  string
	}{
		{
			name:     "data-lazy-src wins over placeholder src",
			imgAttrs: `data-lazy-src="https://charts-static.billboard.com/img/real.jpg" src="https://www.billboard.com/wp-content/themes/vip/pmc-billboard-2021/assets/public/lazyload-fallback.gif"`,
			wantURL:  "https://charts-static.billboard.com/img/real.jpg",
		},
		{
			name:     "plain src works when no lazy attr",
			imgAttrs: `src="https://charts-static.billboard.com/img/real.jpg"`,
			wantURL:  "https://charts-static.billboard.com/img/real.jpg",
		},
		{
			name:     "placeholder-only src yields empty artworkURL",
			imgAttrs: `src="https://www.billboard.com/wp-content/themes/vip/pmc-billboard-2021/assets/public/lazyload-fallback.gif"`,
			wantURL:  "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := extract([]byte(rowHTML(tc.imgAttrs)), 0)
			if err != nil {
				t.Fatalf("extract: %v", err)
			}
			if len(res.Items) == 0 {
				t.Fatal("expected at least one item")
			}
			if got := res.Items[0].ArtworkURL; got != tc.wantURL {
				t.Fatalf("ArtworkURL: want %q got %q", tc.wantURL, got)
			}
		})
	}
}
