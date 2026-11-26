package remoteui

import (
	"context"
	"encoding/json"
	"time"
)

// request asks the desktop for something and hands the reply to done on
// the UI goroutine.
func request[T any](l *link, kind string, body any, done func(T)) {
	s := l.session
	if s == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		var out T
		err := s.Request(ctx, kind, body, &out)
		l.app.post(func() {
			if err != nil {
				l.app.toast(err.Error())
				return
			}
			done(out)
			l.changed()
		})
	}()
}

func decode(data []byte, v any) error { return json.Unmarshal(data, v) }
