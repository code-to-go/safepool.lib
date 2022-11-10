package apps

import (
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/safe"
	"strings"
)

type App interface {
	Receive(safe string, head safe.Head, data []byte, eod bool) bool
	Bind(write func(name string))
}

func AddSafe(s *safe.Safe) {

}

var safes []*safe.Safe

type RoutingRule struct {
	Prefix string
	App    App
}
type Routing []RoutingRule

func Feed(r Routing) {
	for _, s := range safes {
		go feedFrom(s, r)
	}
}

func feedFrom(s *safe.Safe, r Routing) error {
	feedTime, err := sqlGetFeedTime(s.Name)
	if core.IsErr(err, "cannot get feed time: %v") {
		return err
	}

	mostRecent := feedTime
	for _, h := range s.List(0, feedTime) {
		if mostRecent.Before(h.TimeStamp) {
			mostRecent = h.TimeStamp
		}

		var apps []App
		for _, a := range r {
			if strings.HasPrefix(h.Name, a.Prefix) {
				apps = append(apps, a.App)
			}
		}
		err = s.Get(h.Id, &feedWriter{
			safe: s.Name,
			apps: apps,
			head: h,
		})

	}
	return nil
}

type feedWriter struct {
	safe   string
	apps   []App
	head   safe.Head
	offset int64
}

func (w *feedWriter) Write(data []byte) (int, error) {
	w.offset += int64(len(data))
	for _, a := range w.apps {
		if !a.Receive(w.safe, w.head, data, w.offset == w.head.Size) {
			break
		}
	}
	return len(data), nil
}
