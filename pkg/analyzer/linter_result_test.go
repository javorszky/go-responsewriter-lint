package analyzer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLinterResult(t *testing.T) {
	lr := NewLinterResult()

	lr.Record(funcWriteHeader, 9)
	lr.Record(funcWrite, 10)
	lr.Record(funcHeader, 13)
	lr.Record(funcWrite, 23)
	lr.Record(funcWriteHeader, 11)

	assert.Equal(t, map[string][]int{
		funcWrite:       {10, 23},
		funcHeader:      {13},
		funcWriteHeader: {9, 11},
	}, lr.calls)
}

func Test_linterResult_Errors(t *testing.T) {
	tests := []struct {
		name  string
		calls map[string][]int
		want  []error
	}{
		{
			name: "no issues",
			calls: map[string][]int{
				funcHeader:      {5, 8, 9},
				funcWriteHeader: {10},
				funcWrite:       {11},
			},
			want: []error{},
		},
		{
			name: "multiple writes",
			calls: map[string][]int{
				funcWrite: {5, 6},
			},
			want: []error{
				errors.New("there are multiple calls to http.ResponseWriter.Write on the following lines: 5, 6"),
			},
		},
		{
			name: "multiple writeheaders",
			calls: map[string][]int{
				funcWriteHeader: {10, 11},
			},
			want: []error{
				errors.New("there are multiple calls to http.ResponseWriter.WriteHeader on the following lines: 10, 11"),
			},
		},
		{
			name: "writeheader after write",
			calls: map[string][]int{
				funcWriteHeader: {14},
				funcWrite:       {10},
			},
			want: []error{
				errors.New("WriteHeader (line 14) is called after Write (10). Headers are already sent, this has no effect"),
			},
		},
		{
			name: "header after write",
			calls: map[string][]int{
				funcHeader: {10},
				funcWrite:  {8},
			},
			want: []error{
				errors.New("call to w.Header on line 10 comes after a call to w.Write on line 8. This has no effect"),
			},
		},
		{
			name: "header after writeheader",
			calls: map[string][]int{
				funcWriteHeader: {32},
				funcHeader:      {87},
			},
			want: []error{
				errors.New("call to w.Header on line 87 comes after a call to w.WriteHeader on line 32. This has no effect"),
			},
		},
		{
			name: "smörgåsbord",
			calls: map[string][]int{
				funcWrite:       {30},
				funcWriteHeader: {43},
				funcHeader:      {28, 34, 45, 48, 90},
			},
			want: []error{
				errors.New("WriteHeader (line 43) is called after Write (30). Headers are already sent, this has no effect"),
				errors.New("call to w.Header on line 45 comes after a call to w.WriteHeader on line 43. This has no effect"),
				errors.New("call to w.Header on line 48 comes after a call to w.WriteHeader on line 43. This has no effect"),
				errors.New("call to w.Header on line 90 comes after a call to w.WriteHeader on line 43. This has no effect"),
				errors.New("call to w.Header on line 34 comes after a call to w.Write on line 30. This has no effect"),
				errors.New("call to w.Header on line 45 comes after a call to w.Write on line 30. This has no effect"),
				errors.New("call to w.Header on line 48 comes after a call to w.Write on line 30. This has no effect"),
				errors.New("call to w.Header on line 90 comes after a call to w.Write on line 30. This has no effect"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lr := &linterResult{
				calls: tt.calls,
			}

			assert.ElementsMatch(t, tt.want, lr.Errors())
		})
	}
}
