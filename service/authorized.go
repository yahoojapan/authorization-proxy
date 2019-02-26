package service

import authorizationd "github.com/yahoojapan/athenz-policy-updater"

type Authorizationd interface {
	authorizationd.Providerd
}
