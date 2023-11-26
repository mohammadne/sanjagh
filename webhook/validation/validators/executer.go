package validators

import (
	"context"
	"encoding/json"

	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/mohammadne/sanjagh/api/v1alpha1"
	"github.com/mohammadne/sanjagh/webhook/validation/config"
	"github.com/mohammadne/sanjagh/webhook/validation/failure"
)

type executerValidator struct {
	config *config.Config
	client client.Reader
}

func NewExecuter(cfg *config.Config, client client.Reader) *executerValidator {
	return &executerValidator{config: cfg, client: client}
}

func (v *executerValidator) Validate(ctx context.Context, ar *admissionv1.AdmissionReview) (*failure.Failure, error) {
	executer, failure := &v1alpha1.Executer{}, &failure.Failure{}
	if err := json.Unmarshal(ar.Request.Object.Raw, executer); err != nil {
		return nil, err
	}
	if executer.DeletionTimestamp != nil {
		return nil, nil
	}

	if err := v.ValidateReplication(ctx, executer, failure); err != nil {
		return nil, err
	}

	return failure, nil
}

const (
	LowReplication  string = "Replication is lower than the minimum value: '%d'"
	HighReplication string = "Replication exceeds the maximum value: '%d'"
)

func (v *executerValidator) ValidateReplication(ctx context.Context, executer *v1alpha1.Executer, f *failure.Failure) error {
	if executer.Spec.Replication < v.config.MinReplication {
		f.RegisterReason(LowReplication, v.config.MinReplication)
		return nil
	}

	if executer.Spec.Replication > v.config.MaxReplication {
		f.RegisterReason(HighReplication, v.config.MaxReplication)
		return nil
	}

	return nil
}
