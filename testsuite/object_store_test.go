// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package tardigrade_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"storj.io/common/memory"
	"storj.io/common/testcontext"
	"storj.io/common/testrand"
	"storj.io/linksharing/httpserver"
	"storj.io/linksharing/linksharing"
	"storj.io/storj/private/testplanet"
	"storj.io/velero-plugin/tardigrade"
)

func TestPutObject(t *testing.T) {
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

func TestCreateSignedURL(t *testing.T) {
	testplanet.Run(t, testplanet.Config{
		SatelliteCount: 1, StorageNodeCount: 0, UplinkCount: 1,
	}, func(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) {
		linksharingServer, err := newLinksharingServer()
		require.NoError(t, err)

		runCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		ctx.Go(func() error {
			return linksharingServer.Run(runCtx)
		})

		uplink := planet.Uplinks[0]
		satellite := planet.Satellites[0]

		access := uplink.Access[satellite.ID()]
		serializedAccess, err := access.Serialize()
		require.NoError(t, err)

		config := map[string]string{
			"accessGrant": serializedAccess,
		}

		objectStore := tardigrade.NewObjectStore(logrus.New())
		objectStore.LinksharingBaseURL = "http://" + linksharingServer.Addr()
		err = objectStore.Init(config)
		require.NoError(t, err)

		err = uplink.CreateBucket(ctx, satellite, "bucket")
		require.NoError(t, err)

		data := testrand.Bytes(1 * memory.KiB)

		err = uplink.Upload(ctx, satellite, "bucket", "object", data)
		require.NoError(t, err)

		// Create signed URL with TTL of 1 minute
		signedURL, err := objectStore.CreateSignedURL("bucket", "object", 1*time.Minute)
		require.NoError(t, err)

		resp, err := http.Get(signedURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		downloaded, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, data, downloaded)

		// Create signed URL with TTL ending in the past - expect error on download
		signedURL, err = objectStore.CreateSignedURL("bucket", "object", -1*time.Minute)
		require.NoError(t, err)

		resp, err = http.Get(signedURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		downloaded, err = ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})
}

func newLinksharingServer() (*httpserver.Server, error) {
	handler, err := linksharing.NewHandler(zap.NewNop(), linksharing.HandlerConfig{URLBase: "http://localhost:0"})
	if err != nil {
		return nil, err
	}

	server, err := httpserver.New(zap.NewNop(), httpserver.Config{
		Name:            "Link Sharing",
		Address:         "localhost:0",
		Handler:         handler,
		ShutdownTimeout: -1,
	})
	if err != nil {
		return nil, err
	}

	return server, nil
}
