package mylog

import (
	"github.com/pkg/errors"
	"github.com/sweemingdow/gmicro_pkg/pkg/myerr"
)

var (
	_skipFrames    = 3
	_extractFrames = 6
)

func setSkipFrames(frames int) {
	if frames > 0 {
		_skipFrames = frames
	}
}

func setExtractFrames(frames int) {
	if frames > 0 {
		_extractFrames = frames
	}
}

func MarshalStackLimited(err error) interface{} {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	var stackFrames = -1
	cme, ok := myerr.AsCodeMsgError(err)
	if ok {
		stackFrames = _extractFrames
		//err = cme.Floor()
		err = errors.Unwrap(cme)
	}

	var sterr stackTracer
	for err != nil {
		if st, ok := err.(stackTracer); ok {
			sterr = st
			break
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = u.Unwrap()
	}

	if sterr == nil {
		return nil
	}

	st := sterr.StackTrace()
	s := &state{}

	if stackFrames == -1 {
		out := make([]map[string]string, 0, len(st))
		for _, frame := range st {
			out = append(out, map[string]string{
				"file":     frameField(frame, s, 's'),
				"line":     frameField(frame, s, 'd'),
				"function": frameField(frame, s, 'n'),
			})
		}

		return out
	} else {
		frames := st
		if len(frames) > stackFrames {
			frames = frames[_skipFrames:stackFrames]
		}
		out := make([]map[string]string, 0, len(frames))
		for _, frame := range frames {
			out = append(out, map[string]string{
				"file":     frameField(frame, s, 's'),
				"line":     frameField(frame, s, 'd'),
				"function": frameField(frame, s, 'n'),
			})
		}

		return out
	}
}

type state struct {
	b []byte
}

// Write implement fmt.Formatter interface.
func (s *state) Write(b []byte) (n int, err error) {
	s.b = b
	return len(b), nil
}

// Width implement fmt.Formatter interface.
func (s *state) Width() (wid int, ok bool) {
	return 0, false
}

// Precision implement fmt.Formatter interface.
func (s *state) Precision() (prec int, ok bool) {
	return 0, false
}

// Flag implement fmt.Formatter interface.
func (s *state) Flag(c int) bool {
	return false
}

func frameField(f errors.Frame, s *state, c rune) string {
	f.Format(s, c)
	return string(s.b)
}
