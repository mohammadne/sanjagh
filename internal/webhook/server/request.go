package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
)

const (
	RequestContentTypeKey = "Content-Type"
	RequestContentType    = "application/json"
)

func parseRequest(r http.Request) (*admissionv1.AdmissionReview, error) {
	if r.Header.Get(RequestContentTypeKey) != RequestContentType {
		return nil, fmt.Errorf("Content-Type: %q should be %q",
			r.Header.Get(RequestContentTypeKey), RequestContentType)
	}

	bodybuf := new(bytes.Buffer)
	if _, err := bodybuf.ReadFrom(r.Body); err != nil {
		return nil, err
	}
	body := bodybuf.Bytes()

	if len(body) == 0 {
		return nil, fmt.Errorf("admission request body is empty")
	}

	a := admissionv1.AdmissionReview{}

	if err := json.Unmarshal(body, &a); err != nil {
		return nil, fmt.Errorf("could not parse admission review request: %w", err)
	}

	if a.Request == nil {
		return nil, fmt.Errorf("admission review can't be used: Request field is nil")
	}

	return &a, nil
}
