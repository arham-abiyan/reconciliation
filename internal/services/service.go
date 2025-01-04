package services

import "github.com/arham-abiyan/reconciliation/internal/model"

type Reconciliation interface {
	Reconcile() (model.ReconcileResponse, error)
}
