package p

import (
	"net/http"
)

type bla struct{}

func foo(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("some header", "value")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`boys in the yard`))
}

func (b bla) method(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("some header", "value")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`boys in the yard`))
}

func (b *bla) methodPointer(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("some header", "value")
	w.Write([]byte(`boys in the yard`))
	w.WriteHeader(http.StatusOK) // want "function methodPointer: http.ResponseWriter.Write is called before http.ResponseWriter.WriteHeader. Headers are already sent, this has no effect."
}

func bad(bloe http.ResponseWriter, r *http.Request) {
	bloe.Write([]byte(`hellyea`)) // want "function bad: Multiple calls to http.ResponseWriter.Write in the same function body. This is most probably a bug."

	bloe.WriteHeader(http.StatusBadRequest)          // want "function bad: Multiple calls to http.ResponseWriter.WriteHeader in the same function body. This is most probably a bug."
	bloe.Write([]byte(`hellyelamdflmda`))            // want "function bad: Multiple calls to http.ResponseWriter.Write in the same function body. This is most probably a bug."
	bloe.WriteHeader(http.StatusInternalServerError) // want "function bad: Multiple calls to http.ResponseWriter.WriteHeader in the same function body. This is most probably a bug."

	bloe.Header().Set("help", "somebody")     // want "function bad: http.ResponseWriter.Header called after calling http.ResponseWriter.Write. This has no effect." "function bad: http.ResponseWriter.Header called after calling http.ResponseWriter.Write. This has no effect." "function bad: http.ResponseWriter.Header called after calling http.ResponseWriter.WriteHeader. This has no effect." "function bad: http.ResponseWriter.Header called after calling http.ResponseWriter.WriteHeader. This has no effect."
	bloe.Header().Set("dddd", "someboddaady") // want "function bad: http.ResponseWriter.Header called after calling http.ResponseWriter.Write. This has no effect." "function bad: http.ResponseWriter.Header called after calling http.ResponseWriter.Write. This has no effect." "function bad: http.ResponseWriter.Header called after calling http.ResponseWriter.WriteHeader. This has no effect." "function bad: http.ResponseWriter.Header called after calling http.ResponseWriter.WriteHeader. This has no effect."
}
