package analyzer

import (
	"go/token"
)

const (
	funcHeader      = "Header"
	funcWrite       = "Write"
	funcWriteHeader = "WriteHeader"
)

type linterError struct {
	message string
	pos     token.Pos
}

// linterResult stores the line numbers of function calls within a function / method for the individual methods we're
// looking for.
type linterResult struct {
	fn    string
	calls map[string][]position
}

type position struct {
	line int
	pos  token.Pos
}

func NewLinterResult(fn string) *linterResult {
	return &linterResult{
		fn: fn,
		calls: map[string][]position{
			funcHeader:      make([]position, 0),
			funcWrite:       make([]position, 0),
			funcWriteHeader: make([]position, 0),
		},
	}
}

func (lr *linterResult) Record(funcName string, line int, pos token.Pos) {
	lr.calls[funcName] = append(lr.calls[funcName], position{
		line: line,
		pos:  pos,
	})
}

func (lr *linterResult) Errors() []linterError {
	errs := make([]linterError, 0)

	// Multiple calls to http.ResponseWriter.Write([]byte).
	if len(lr.calls[funcWrite]) > 1 {
		for _, l := range lr.calls[funcWrite] {
			errs = append(errs, linterError{
				message: "Multiple calls to http.ResponseWriter.Write in the same function body. This is most probably a bug.",
				pos:     l.pos,
			})
		}
	}

	// Multiple calls to http.ResponseWriter.WriteHeader(int).
	if len(lr.calls[funcWriteHeader]) > 1 {
		for _, l := range lr.calls[funcWriteHeader] {
			errs = append(errs, linterError{
				message: "Multiple calls to http.ResponseWriter.WriteHeader in the same function body. This is most probably a bug.",
				pos:     l.pos,
			})
		}
	}

	// Calling w.WriteHeader() after w.Write()
	if len(lr.calls[funcWrite]) == 1 && len(lr.calls[funcWriteHeader]) == 1 {
		writeLine := lr.calls[funcWrite][0]
		whLine := lr.calls[funcWriteHeader][0]
		if whLine.line > writeLine.line {
			errs = append(errs,
				linterError{
					message: "http.ResponseWriter.Write is called before http.ResponseWriter.WriteHeader. Headers are already sent, this has no effect.",
					pos:     whLine.pos,
				},
			)
		}
	}

	// Calling w.Header() after w.WriteHeader() or w.Write()
	for _, hl := range lr.calls[funcHeader] {
		for _, wl := range lr.calls[funcWrite] {
			if hl.line > wl.line {
				errs = append(errs, linterError{
					message: "http.ResponseWriter.Header called after calling http.ResponseWriter.Write. This has no effect.",
					pos:     hl.pos,
				})
			}
		}

		for _, whl := range lr.calls[funcWriteHeader] {
			if hl.line > whl.line {
				errs = append(errs,
					linterError{
						message: "http.ResponseWriter.Header called after calling http.ResponseWriter.WriteHeader. This has no effect.",
						pos:     hl.pos,
					},
				)
			}
		}
	}

	return errs
}
