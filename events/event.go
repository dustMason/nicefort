package events

import (
	"strings"
	"time"
)
import "container/list"

type Class int

const (
	Warning Class = iota
	Success
	Danger
	Info
)

type Event struct {
	kind    Class
	text    string
	subject string // optional element
	when    time.Time
}

func (e Event) render(b *strings.Builder) {
	// todo styling based on kind
	if e.subject != "" {
		b.WriteString(e.subject + ": ")
	}
	b.WriteString(e.text)
}

type EventList struct {
	len int
	list.List
	// todo add option to render timestamps?
}

func NewEventList(len int) *EventList {
	return &EventList{len: len}
}

func (el *EventList) Add(kind Class, text string) {
	el.AddWithSubject(kind, text, "")
}

func (el *EventList) AddWithSubject(kind Class, text, subject string) {
	e := Event{
		kind:    kind,
		text:    text,
		subject: subject,
		when:    time.Now(),
	}
	el.PushFront(e)
	for el.Len() > el.len {
		last := el.Back()
		el.Remove(last)
	}
}

func (el *EventList) Render() string {
	var b strings.Builder
	n := el.Back()
	for n != nil {
		// out = append(out, n.Value.(Event))
		n.Value.(Event).render(&b)
		b.WriteString("\n")
		n = n.Prev()
	}
	// for _, e := range el.Events() {
	// 	e.render(&b)
	// 	b.WriteString("\n")
	// }
	return b.String()
}

// func (el *EventList) Events() []Event {
// 	out := make([]Event, 0)
// 	n := el.Back()
// 	for n != nil {
// 		out = append(out, n.Value.(Event))
// 		n = n.Prev()
// 	}
// 	return out
// }
