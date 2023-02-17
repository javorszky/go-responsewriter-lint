package analyzer

import (
	"go/token"
)

const (
	FuncHeader      = "Header"
	FuncWrite       = "Write"
	FuncWriteHeader = "WriteHeader"
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
			FuncHeader:      make([]position, 0),
			FuncWrite:       make([]position, 0),
			FuncWriteHeader: make([]position, 0),
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
	if len(lr.calls[FuncWrite]) > 1 {
		for _, l := range lr.calls[FuncWrite] {
			errs = append(errs, linterError{
				message: "Multiple calls to http.ResponseWriter.Write in the same function body. This is most probably a bug.",
				pos:     l.pos,
			})
		}
	}

	// Multiple calls to http.ResponseWriter.WriteHeader(int).
	if len(lr.calls[FuncWriteHeader]) > 1 {
		for _, l := range lr.calls[FuncWriteHeader] {
			errs = append(errs, linterError{
				message: "Multiple calls to http.ResponseWriter.WriteHeader in the same function body. This is most probably a bug.",
				pos:     l.pos,
			})
		}
	}

	// Calling w.WriteHeader() after w.Write()
	if len(lr.calls[FuncWrite]) == 1 && len(lr.calls[FuncWriteHeader]) == 1 {
		writeLine := lr.calls[FuncWrite][0]
		whLine := lr.calls[FuncWriteHeader][0]
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
	for _, hl := range lr.calls[FuncHeader] {
		for _, wl := range lr.calls[FuncWrite] {
			if hl.line > wl.line {
				errs = append(errs, linterError{
					message: "http.ResponseWriter.Header called after calling http.ResponseWriter.Write. This has no effect.",
					pos:     hl.pos,
				})
			}
		}

		for _, whl := range lr.calls[FuncWriteHeader] {
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
