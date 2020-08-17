// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	veleroplugin "github.com/vmware-tanzu/velero/pkg/plugin/framework"

	"storj.io/velero-plugin/tardigrade"
)

func main() {
	veleroplugin.NewServer().
		BindFlags(pflag.CommandLine).
		RegisterObjectStore("velero.io/tardigrade", newObjectStore).
		Serve()
}

func newObjectStore(logger logrus.FieldLogger) (interface{}, error) {
	return tardigrade.NewObjectStore(logger), nil
}
