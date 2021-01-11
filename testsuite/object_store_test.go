// Copyright (C) 2020 Storj Labs, Inc.
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
	"storj.io/linksharing/sharing"
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

func TestObjectExists(t *testing.T) {
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

		// Exect false if bucket does not exist
		exists, err := objectStore.ObjectExists("bucket", "object")
		require.NoError(t, err)
		assert.False(t, exists)

		err = uplink.CreateBucket(ctx, satellite, "bucket")
		require.NoError(t, err)

		// Expect false if object does not exist
		exists, err = objectStore.ObjectExists("bucket", "object")
		require.NoError(t, err)
		assert.False(t, exists)

		err = uplink.Upload(ctx, satellite, "bucket", "object", testrand.Bytes(1*memory.KiB))
		require.NoError(t, err)

		exists, err = objectStore.ObjectExists("bucket", "object")
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestGetObject(t *testing.T) {
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

		// Expect error if bucket does not exist
		_, err = objectStore.GetObject("bucket", "object")
		require.Error(t, err)

		err = uplink.CreateBucket(ctx, satellite, "bucket")
		require.NoError(t, err)

		// Expect error if object does not exist
		_, err = objectStore.GetObject("bucket", "object")
		require.Error(t, err)

		data := testrand.Bytes(1 * memory.KiB)

		err = uplink.Upload(ctx, satellite, "bucket", "object", data)
		require.NoError(t, err)

		reader, err := objectStore.GetObject("bucket", "object")
		require.NoError(t, err)

		downloaded, err := ioutil.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, data, downloaded)
	})
}

func TestListCommonPrefixes(t *testing.T) {
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

		// Expect error if bucket does not exist
		_, err = objectStore.ListCommonPrefixes("bucket", "a-prefix/", "/")
		require.Error(t, err)

		err = uplink.CreateBucket(ctx, satellite, "bucket")
		require.NoError(t, err)

		// Expect no error but empty result if the bucket is empty
		items, err := objectStore.ListCommonPrefixes("bucket", "a-prefix/", "/")
		require.NoError(t, err)
		assert.Empty(t, items)

		err = uplink.Upload(ctx, satellite, "bucket", "a-prefix/foo-1/bar", nil)
		require.NoError(t, err)

		err = uplink.Upload(ctx, satellite, "bucket", "a-prefix/foo-1/baz", nil)
		require.NoError(t, err)

		err = uplink.Upload(ctx, satellite, "bucket", "a-prefix/foo-2/baz", nil)
		require.NoError(t, err)

		err = uplink.Upload(ctx, satellite, "bucket", "a-prefix/foo-2/foo/baz", nil)
		require.NoError(t, err)

		err = uplink.Upload(ctx, satellite, "bucket", "a-prefix/bar", nil)
		require.NoError(t, err)

		err = uplink.Upload(ctx, satellite, "bucket", "some-other-prefix/foo-3/bar", nil)
		require.NoError(t, err)

		items, err = objectStore.ListCommonPrefixes("bucket", "a-prefix/", "/")
		require.NoError(t, err)
		assert.Len(t, items, 2)
		assert.Contains(t, items, "a-prefix/foo-1/")
		assert.Contains(t, items, "a-prefix/foo-2/")
	})
}

func TestListObjects(t *testing.T) {
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

		// Expect error if bucket does not exist
		_, err = objectStore.ListObjects("bucket", "a-prefix/")
		require.Error(t, err)

		err = uplink.CreateBucket(ctx, satellite, "bucket")
		require.NoError(t, err)

		// Expect no error but empty result if the bucket is empty
		items, err := objectStore.ListObjects("bucket", "a-prefix/")
		require.NoError(t, err)
		assert.Empty(t, items)

		err = uplink.Upload(ctx, satellite, "bucket", "a-prefix/foo-1/bar", nil)
		require.NoError(t, err)

		err = uplink.Upload(ctx, satellite, "bucket", "a-prefix/foo-1/baz", nil)
		require.NoError(t, err)

		err = uplink.Upload(ctx, satellite, "bucket", "a-prefix/foo-2/baz", nil)
		require.NoError(t, err)

		err = uplink.Upload(ctx, satellite, "bucket", "a-prefix/bar", nil)
		require.NoError(t, err)

		err = uplink.Upload(ctx, satellite, "bucket", "some-other-prefix/foo-3/bar", nil)
		require.NoError(t, err)

		items, err = objectStore.ListObjects("bucket", "a-prefix/")
		require.NoError(t, err)
		assert.Len(t, items, 4)
		assert.Contains(t, items, "a-prefix/foo-1/bar")
		assert.Contains(t, items, "a-prefix/foo-1/baz")
		assert.Contains(t, items, "a-prefix/foo-2/baz")
		assert.Contains(t, items, "a-prefix/bar")
	})
}

func TestDeleteObject(t *testing.T) {
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

		// Expect no error if bucket does not exist
		err = objectStore.DeleteObject("bucket", "object")
		require.NoError(t, err)

		err = uplink.CreateBucket(ctx, satellite, "bucket")
		require.NoError(t, err)

		// Expect no error if object does not exist
		err = objectStore.DeleteObject("bucket", "object")
		require.NoError(t, err)

		data := testrand.Bytes(1 * memory.KiB)

		err = uplink.Upload(ctx, satellite, "bucket", "object", data)
		require.NoError(t, err)

		err = objectStore.DeleteObject("bucket", "object")
		require.NoError(t, err)

		// Check that the object does not exist anymore
		_, err = uplink.Download(ctx, satellite, "bucket", "object")
		require.Error(t, err)
	})
}

func TestCreateSignedURL(t *testing.T) {
	// Skipping test due to incompatible setup of linksharing (missing html files)
	t.Skip()
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
	handler, err := sharing.NewHandler(zap.NewNop(), nil, sharing.Config{URLBase: "http://localhost:0"})
	if err != nil {
		return nil, err
	}

	server, err := httpserver.New(zap.NewNop(), handler, httpserver.Config{
		Name:            "Link Sharing",
		Address:         "localhost:0",
		ShutdownTimeout: -1,
	})
	if err != nil {
		return nil, err
	}

	return server, nil
}
