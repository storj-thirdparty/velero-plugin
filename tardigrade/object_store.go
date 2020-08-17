// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package tardigrade

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	veleroplugin "github.com/vmware-tanzu/velero/pkg/plugin/framework"

	"storj.io/uplink"
)

// Config params.
const (
	accessGrant = "accessGrant"
)

const defaultLinksharingBaseURL = "https://link.tardigradeshare.io"

// ObjectStore exposes basic object-storage operations required
// by Velero.
type ObjectStore struct {
	log                logrus.FieldLogger
	access             *uplink.Access
	project            *uplink.Project
	LinksharingBaseURL string
}

// NewObjectStore creates new Tardigrade object store.
func NewObjectStore(logger logrus.FieldLogger) *ObjectStore {
	return &ObjectStore{log: logger, LinksharingBaseURL: defaultLinksharingBaseURL}
}

// Init prepares the ObjectStore for usage using the provided map of
// configuration key-value pairs. It returns an error if the ObjectStore
// cannot be initialized from the provided config.
//
// Specific config for Tardigrade:
//   - accessGrant (required): serialized access grant to Tardigrade project.
func (o *ObjectStore) Init(config map[string]string) error {
	o.log.Debug("objectStore.Init called")

	err := veleroplugin.ValidateObjectStoreConfigKeys(config, accessGrant)
	if err != nil {
		return err
	}

	o.access, err = uplink.ParseAccess(config[accessGrant])
	if err != nil {
		return err
	}

	o.project, err = uplink.OpenProject(context.Background(), o.access)
	if err != nil {
		return err
	}

	return nil
}

// PutObject creates a new object using the data in body within the specified
// object storage bucket with the given key.
func (o *ObjectStore) PutObject(bucket, key string, body io.Reader) error {
	o.log.Debug("objectStore.PutObject called")

	upload, err := o.project.UploadObject(context.Background(), bucket, key, nil)
	if err != nil {
		return err
	}

	if _, err := io.Copy(upload, body); err != nil {
		return err
	}

	return upload.Commit()
}

// ObjectExists checks if there is an object with the given key in the object storage bucket.
func (o *ObjectStore) ObjectExists(bucket, key string) (bool, error) {
	o.log.Debug("objectStore.ObjectExists called")

	if _, err := o.project.StatObject(context.Background(), bucket, key); err != nil {
		if errors.Is(err, uplink.ErrObjectNotFound) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}

	return true, nil
}

// GetObject retrieves the object with the given key from the specified
// bucket in object storage.
func (o *ObjectStore) GetObject(bucket, key string) (io.ReadCloser, error) {
	o.log.Debug("objectStore.GetObject called")

	downloader, err := o.project.DownloadObject(context.Background(), bucket, key, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return downloader, nil
}

// ListCommonPrefixes gets a list of all object key prefixes that start with
// the specified prefix and stop at the next instance of the provided delimiter.
//
// For example, if the bucket contains the following keys:
//		a-prefix/foo-1/bar
// 		a-prefix/foo-1/baz
//		a-prefix/foo-2/baz
// 		some-other-prefix/foo-3/bar
// and the provided prefix arg is "a-prefix/", and the delimiter is "/",
// this will return the slice {"a-prefix/foo-1/", "a-prefix/foo-2/"}.
func (o *ObjectStore) ListCommonPrefixes(bucket, prefix, delimiter string) ([]string, error) {
	o.log.Debug("objectStore.ListCommonPrefixes called")

	objectsIter := o.project.ListObjects(context.Background(), bucket, &uplink.ListObjectsOptions{Prefix: prefix})
	var res []string
	for objectsIter.Next() {
		object := objectsIter.Item()
		if object.IsPrefix {
			res = append(res, object.Key)
		}
	}

	if err := objectsIter.Err(); err != nil {
		return res, err
	}

	return res, nil
}

// ListObjects gets a list of all keys in the specified bucket
// that have the given prefix.
func (o *ObjectStore) ListObjects(bucket, prefix string) ([]string, error) {
	o.log.Debug("objectStore.ListObjects called")

	object := o.project.ListObjects(context.Background(), bucket, &uplink.ListObjectsOptions{Prefix: prefix, Recursive: true})
	var res []string
	for object.Next() {
		res = append(res, object.Item().Key)
	}

	if err := object.Err(); err != nil {
		return res, err
	}

	return res, nil
}

// DeleteObject removes the object with the specified key from the given
// bucket.
func (o *ObjectStore) DeleteObject(bucket, key string) error {
	o.log.Debug("objectStore.DeleteObject called")

	_, err := o.project.DeleteObject(context.Background(), bucket, key)
	if err != nil {
		if errors.Is(err, uplink.ErrBucketNotFound) {
			return nil
		}
		if errors.Is(err, uplink.ErrObjectNotFound) {
			return nil
		}
		return err
	}

	return nil
}

// CreateSignedURL creates a pre-signed URL for the given bucket and key that expires after ttl.
func (o *ObjectStore) CreateSignedURL(bucket, key string, ttl time.Duration) (string, error) {
	o.log.Debug("objectStore.CreateSignedURL called")

	permission := uplink.ReadOnlyPermission()
	permission.NotAfter = time.Now().Add(ttl)

	restrictedAccess, err := o.access.Share(permission, uplink.SharePrefix{
		Bucket: bucket,
		Prefix: key,
	})
	if err != nil {
		return "", err
	}

	restrictedAccessGrant, err := restrictedAccess.Serialize()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s/%s", o.LinksharingBaseURL,
		url.PathEscape(restrictedAccessGrant),
		url.PathEscape(bucket),
		url.PathEscape(key)), nil
}
