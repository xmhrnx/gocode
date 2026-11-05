package openai

import "strings"

// thinkSplitter separates inline <think>...</think> reasoning, which some
// local servers leave in the content stream, from reply text. Tags may be
// split across chunks, so a possible partial tag is held back until the
// next chunk decides it. An opening tag only counts before any reply text,
// so a reply that merely mentions "<think>" is left alone.
type thinkSplitter struct {
	pending string
	inThink bool
	sawText bool
}

const (
	openThink  = "<think>"
	closeThink = "</think>"
)

// feed processes one chunk, calling emit with each run of reasoning or
// reply text it can already classify.
func (s *thinkSplitter) feed(chunk string, emit func(reasoning bool, text string)) {
	s.pending += chunk
	for s.pending != "" {
		if s.inThink {
			if !s.emitUntil(closeThink, emit) {
				return
			}
			s.inThink = false
			continue
		}
		if s.sawText {
			emit(false, s.pending)
			s.pending = ""
			return
		}
		trimmed := strings.TrimLeft(s.pending, " \t\r\n")
		if strings.HasPrefix(trimmed, openThink) {
			s.pending, s.inThink = trimmed[len(openThink):], true
			continue
		}
		if strings.HasPrefix(openThink, trimmed) {
			return // whitespace or a partial opening tag: wait for more
		}
		// Leading whitespace (often left after "</think>") is dropped.
		s.pending, s.sawText = trimmed, true
	}
}

// emitUntil emits reasoning up to tag and reports whether the tag was
// found; otherwise it keeps any trailing partial tag pending.
func (s *thinkSplitter) emitUntil(tag string, emit func(reasoning bool, text string)) bool {
	if index := strings.Index(s.pending, tag); index >= 0 {
		emit(true, s.pending[:index])
		s.pending = s.pending[index+len(tag):]
		return true
	}
	keep := partialSuffix(s.pending, tag)
	emit(true, s.pending[:len(s.pending)-keep])
	s.pending = s.pending[len(s.pending)-keep:]
	return false
}

// flush emits whatever is still held back at the end of the stream.
func (s *thinkSplitter) flush(emit func(reasoning bool, text string)) {
	if s.pending != "" {
		emit(s.inThink, s.pending)
		s.pending = ""
	}
}

// partialSuffix returns the length of the longest suffix of text that is a
// proper prefix of tag.
func partialSuffix(text, tag string) int {
	for n := min(len(tag)-1, len(text)); n > 0; n-- {
		if strings.HasSuffix(text, tag[:n]) {
			return n
		}
	}
	return 0
}
