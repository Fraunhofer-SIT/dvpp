package messages

import (
	"encoding/json"
	"encoding/xml"
)

type HTTPRequest interface {
	ToPath() string
}

type Result struct {
	Success  bool     `json:"success" xml:"success"`
	Response string   `json:"response" xml:"response"`
	Errors   []string `json:"errors" xml:"errors>error"`
}

func (r *Result) ToJSON() []byte {
	json, _ := json.MarshalIndent(r, "", "\t")
	return json
}

func (r *Result) ToXML() []byte {
	xml, _ := xml.MarshalIndent(
		struct {
			Result
			XMLName struct{} `xml:"result"`
		}{Result: *r},
		"",
		"\t",
	)
	return xml
}

func (r *Result) AppendError(err string) {
	r.Errors = append(r.Errors, err)
}

func (r *Result) AppendErrorWithPrefix(err string, prefix string) {
	r.AppendError(prefix + ": " + err)
}

func (r *Result) AppendErrors(errs []string) {
	for _, err := range errs {
		r.AppendError(err)
	}
}

func (r *Result) AppendErrorsWithPrefix(errs []string, prefix string) {
	for _, err := range errs {
		r.AppendErrorWithPrefix(err, prefix)
	}
}

func NewResult() Result {
	return Result{false, "", []string{}}
}
