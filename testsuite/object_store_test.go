// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package tardigrade_test

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"storj.io/common/memory"
	"storj.io/common/testcontext"
	"storj.io/common/testrand"
	"storj.io/storj/private/testplanet"
	"storj.io/velero-plugin/tardigrade"
)

func TestUploadDownload(t *testing.T) {
	testplanet.Run(t, testplanet.Config{
		SatelliteCount: 1, StorageNodeCount: 0, UplinkCount: 1,
	}, func(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) {
		uplink := planet.Uplinks[0]
		satellite := planet.Satellites[0]

		access := uplink.Access[satellite.ID()]
		serializedAccess, err := access.Serialize()
		require.NoError(t, err)

		config := map[string]string{
			"accessGrant": serializedAccess,
		}

		objectStore := tardigrade.NewObjectStore(logrus.New())
		err = objectStore.Init(config)
		require.NoError(t, err)

		err = uplink.CreateBucket(ctx, satellite, "bucket")
		require.NoError(t, err)

		data := testrand.Bytes(1 * memory.KiB)

		err = objectStore.PutObject("bucket", "object", bytes.NewBuffer(data))
		require.NoError(t, err)

		downloaded, err := uplink.Download(ctx, satellite, "bucket", "object")
		require.NoError(t, err)

		assert.Equal(t, data, downloaded)
	})
}
