package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
)

const (
	ResponseContentTypeKey = "Content-Type"
	ResponseContentType    = "application/json"
)

func writeResponse(ar *admissionv1.AdmissionReview, w http.ResponseWriter) error {
	jout, err := json.Marshal(ar)
	if err != nil {
		return fmt.Errorf("could not marshal admission review: %w", err)
	}

	w.Header().Set(ResponseContentTypeKey, ResponseContentType)
	if _, err := fmt.Fprintf(w, "%s", jout); err != nil {
		return fmt.Errorf("could not write response: %w", err)
	}
	return nil
}
