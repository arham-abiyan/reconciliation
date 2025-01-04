package services

import "github.com/arham09/reconciliation-svc/internal/model"

type Reconciliation interface {
	Reconcile() (model.ReconcileResponse, error)
}
