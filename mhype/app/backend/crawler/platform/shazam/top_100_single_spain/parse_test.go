package top_100_single_spain

import (
	"fmt"
	"strings"
	"testing"
)

func TestExtractFromCSV(t *testing.T) {
	csvFixture := "\xef\xbb\xbf\"Monday, 11 May 2026 [performance over the past 7 days]\"\nRank,Artist,Title\n1,Bad Bunny,DTMF\n2,Shakira,La Bicicleta\n3,Rosalía,Malamente\n"

	result, err := extract([]byte(csvFixture), nil, 1715385600000)
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if !result.Snapshot {
		t.Error("expected Snapshot == true")
	}
	if result.BundleKey != "current" {
		t.Errorf("BundleKey = %q, want %q", result.BundleKey, "current")
	}
	if len(result.Items) != 3 {
		t.Fatalf("got %d items, want 3", len(result.Items))
	}

	cases := []struct {
		rank   int
		title  string
		artist string
	}{
		{1, "DTMF", "Bad Bunny"},
		{2, "La Bicicleta", "Shakira"},
		{3, "Malamente", "Rosalía"},
	}
	for i, c := range cases {
		item := result.Items[i]
		if item.Chart.Position != c.rank {
			t.Errorf("items[%d].Position = %d, want %d", i, item.Chart.Position, c.rank)
		}
		if item.Title != c.title {
			t.Errorf("items[%d].Title = %q, want %q", i, item.Title, c.title)
		}
		if len(item.Artists) == 0 || item.Artists[0] != c.artist {
			t.Errorf("items[%d].Artists[0] = %v, want %q", i, item.Artists, c.artist)
		}
		if item.ArtworkURL != "" {
			t.Errorf("items[%d].ArtworkURL = %q, want empty (no HTML provided)", i, item.ArtworkURL)
		}
	}

	if result.Items[0].Chart.PeriodEnd != "2026-05-11" {
		t.Errorf("PeriodEnd = %q, want 2026-05-11", result.Items[0].Chart.PeriodEnd)
	}
	if result.Items[0].Chart.PeriodStart != "2026-05-04" {
		t.Errorf("PeriodStart = %q, want 2026-05-04", result.Items[0].Chart.PeriodStart)
	}
}

func TestExtractMergesArtwork(t *testing.T) {
	csvFixture := "Rank,Artist,Title\n1,Bad Bunny,DTMF\n"
	// Real DOM structure: class lives on <div>, not <li>. The <li> has empty class.
	// A non-mzstatic sprite img appears first inside the card; the parser must skip it
	// and pick the mzstatic.com album-art image instead (CDN filter).
	htmlFixture := `<html><body><ul>
<li class="">
  <div class="page_songItem__lAdHy">
    <span class="SongItem-module_rankingNumber__abc123">1</span>
    <img srcset="https://assets.shazam.com/sprite.webp 100w,https://assets.shazam.com/sprite.webp 300w">
    <img srcset="https://is1-ssl.mzstatic.com/image/75.webp 75w,https://is1-ssl.mzstatic.com/image/375.webp 375w,https://is1-ssl.mzstatic.com/image/150.webp 150w">
  </div>
</li>
</ul></body></html>`

	result, err := extract([]byte(csvFixture), []byte(htmlFixture), 1715385600000)
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(result.Items))
	}
	want := "https://is1-ssl.mzstatic.com/image/375.webp"
	if result.Items[0].ArtworkURL != want {
		t.Errorf("ArtworkURL = %q, want %q", result.Items[0].ArtworkURL, want)
	}
}

func TestExtractLimitsTo100(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("Rank,Artist,Title\n")
	for i := 1; i <= 150; i++ {
		sb.WriteString(fmt.Sprintf("%d,Artist%d,Title%d\n", i, i, i))
	}

	result, err := extract([]byte(sb.String()), nil, 1715385600000)
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if len(result.Items) != 100 {
		t.Errorf("got %d items, want 100", len(result.Items))
	}
	if result.Items[99].Chart.Position != 100 {
		t.Errorf("last item position = %d, want 100", result.Items[99].Chart.Position)
	}
}
