package crawler

type CrawledLinks map[string]bool

func (v CrawledLinks) List() []string {
	links := make([]string, 0)
	for s := range v {
		links = append(links, s)
	}

	return links
}
