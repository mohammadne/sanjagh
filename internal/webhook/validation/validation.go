package validation

import (
	"context"
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/mohammadne/sanjagh/internal/webhook/validation/config"
	"github.com/mohammadne/sanjagh/internal/webhook/validation/failure"
	"github.com/mohammadne/sanjagh/internal/webhook/validation/validators"
)

type Validation interface {
	Validate(context.Context, *admissionv1.AdmissionReview) error
}

func NewValidation(cfg *config.Config, client crclient.Reader) Validation {
	v := &validation{client: client}

	// add more validators here
	v.executersValidator = validators.NewExecuter(cfg, client).Validate

	return v
}

type Validator func(context.Context, *admissionv1.AdmissionReview) (*failure.Failure, error)

type validation struct {
	client client.Reader

	executersValidator Validator
}

func (v *validation) Validate(ctx context.Context, ar *admissionv1.AdmissionReview) error {
	var failure *failure.Failure
	var err error

	switch ar.Request.Resource.Resource {
	case "executers":
		failure, err = v.executersValidator(ctx, ar)
	default:
		err = fmt.Errorf("unsupported resource: %s", ar.Request.Resource.Resource)
	}

	if err != nil {
		return err
	}

	// generate response
	ar.Response = &admissionv1.AdmissionResponse{
		UID:     ar.Request.UID,
		Allowed: failure.IsAllowed(),
		Result:  &metav1.Status{Message: failure.Reason()},
	}

	return nil
}
