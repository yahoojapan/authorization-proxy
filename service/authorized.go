package service

import authorizationd "github.com/yahoojapan/athenz-policy-updater"

type Authorization interface {
	authorizationd.Providerd
}
