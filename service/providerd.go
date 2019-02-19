package service

import providerd "github.com/yahoojapan/athenz-policy-updater"

type Provider interface {
	providerd.Providerd
}
