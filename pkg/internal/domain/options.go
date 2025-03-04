package domain

import "gorm.io/gorm/clause"

type VapiReleaseDFSOption struct {
	skipVisited bool
}

type VapiReleaseDFSOptionFunc func(*VapiReleaseDFSOption)

func SkipVisited() VapiReleaseDFSOptionFunc {
	return func(o *VapiReleaseDFSOption) {
		o.skipVisited = true
	}
}

func mergeVapiReleaseDFSOptions(options ...VapiReleaseDFSOptionFunc) VapiReleaseDFSOption {
	var opt VapiReleaseDFSOption
	for _, o := range options {
		o(&opt)
	}
	return opt
}

type FindOption struct {
	ignoreErrorOnNotFound bool
	publicOnly            bool
	locking               *clause.Locking
}

type FindOptions func(*FindOption)

func IgnoreErrorOnNotFound() FindOptions {
	return func(o *FindOption) {
		o.ignoreErrorOnNotFound = true
	}
}

func PublicOnly() FindOptions {
	return func(o *FindOption) {
		o.publicOnly = true
	}
}

func Locking(locking clause.Locking) FindOptions {
	return func(o *FindOption) {
		o.locking = &locking
	}
}

func MergeFindOptions(options ...FindOptions) FindOption {
	var opt FindOption
	for _, o := range options {
		o(&opt)
	}
	return opt
}
