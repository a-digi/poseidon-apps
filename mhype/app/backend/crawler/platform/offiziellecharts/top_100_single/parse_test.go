package top_100_single

import (
	"os"
	"testing"
	"time"

	"mhype-plugin/crawler"
)

func TestExtractFixture(t *testing.T) {
	body, err := os.ReadFile("/tmp/oc.html")
	if err != nil {
		t.Skipf("fixture /tmp/oc.html not present: %v", err)
	}
	res, err := extract(body, time.Now().UnixMilli())
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if res.BundleKey == "" {
		t.Fatal("bundle key empty")
	}
	if got := len(res.Items); got != 100 {
		t.Fatalf("want 100 items, got %d", got)
	}
	for i, it := range res.Items {
		if err := it.Validate(); err != nil {
			t.Fatalf("item[%d] invalid: %v", i, err)
		}
		if it.Kind != crawler.KindChartEntry {
			t.Fatalf("item[%d] wrong kind: %s", i, it.Kind)
		}
		if it.Chart.Position != i+1 {
			t.Fatalf("item[%d] position: want %d got %d", i, i+1, it.Chart.Position)
		}
	}
}

func TestSplitArtists(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"Justin Bieber feat. Nicki Minaj", []string{"Justin Bieber", "Nicki Minaj"}},
		{"Dardan / Azet", []string{"Dardan", "Azet"}},
		{"Calvin Harris & Disciples", []string{"Calvin Harris", "Disciples"}},
		{"Drake ft. Future", []string{"Drake", "Future"}},
		{"A x B", []string{"A", "B"}},
		{"Single Artist", []string{"Single Artist"}},
	}
	for _, c := range cases {
		got := splitArtists(c.in)
		if len(got) != len(c.want) {
			t.Fatalf("%q: want %v got %v", c.in, c.want, got)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("%q[%d]: want %q got %q", c.in, i, c.want[i], got[i])
			}
		}
	}
}
