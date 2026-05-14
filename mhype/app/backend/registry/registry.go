package registry

import (
	"mhype-plugin/crawler"
	applemusicUKTop100 "mhype-plugin/crawler/platform/applemusic/uk_top_100"
	billboardHot100 "mhype-plugin/crawler/platform/billboard/hot_100"
	offiziellechartsTop100Single "mhype-plugin/crawler/platform/offiziellecharts/top_100_single"
	shazamTop100SingleSpain "mhype-plugin/crawler/platform/shazam/top_100_single_spain"
)

// Registered returns the list of all crawlers the composer should run.
// To register a new crawler, append its New() constructor to the slice.
// Crawlers are listed in display order (first crawler is rendered first
// in the UI's CrawlerCard list).
func Registered() []crawler.Crawler {
	return []crawler.Crawler{
		offiziellechartsTop100Single.New(),
		applemusicUKTop100.New(),
		billboardHot100.New(),
		shazamTop100SingleSpain.New(),
	}
}
