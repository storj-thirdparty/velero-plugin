package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"storj.io/uplink"

	veleroplugin "github.com/vmware-tanzu/velero/pkg/plugin/framework"
)

const (
	accessGrant        = "accessGrant"
	linksharingBaseURL = "https://link.tardigradeshare.io"
)

type ObjectStore struct {
	log     logrus.FieldLogger
	project *uplink.Project
	access  *uplink.Access
}

func newObjectStore(logger logrus.FieldLogger) *ObjectStore {
	return &ObjectStore{log: logger}
}

func setupUplink(ctx context.Context, storjAccessGrant string) (*uplink.Project, error) {
	access, err := uplink.ParseAccess(storjAccessGrant)
	if err != nil {
		return nil, err
	}
	project, err := uplink.OpenProject(ctx, access)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (o *ObjectStore) Init(config map[string]string) error {
	o.log.Infof("objectStore.Init called")
	if err := veleroplugin.ValidateObjectStoreConfigKeys(config, accessGrant); err != nil {
		return err
	}
	access, err := uplink.ParseAccess(config[accessGrant])
	if err != nil {
		return err
	}
	o.access = access
	project, err := setupUplink(context.Background(), config[accessGrant])
	if err != nil {
		return err
	}
	o.project = project
	return nil
}

func (o *ObjectStore) PutObject(bucket, key string, body io.Reader) error {
	o.log.Infof("objectStore.PutObject called")
	upload, err := o.project.UploadObject(context.Background(), bucket, key, nil)
	if err != nil {
		return err
	}
	if _, err := io.Copy(upload, body); err != nil {
		return err
	}
	return upload.Commit()
}

func (o *ObjectStore) ObjectExists(bucket, key string) (bool, error) {
	o.log.Infof("objectStore.ObjectExists called")
	if _, err := o.project.StatObject(context.Background(), bucket, key); err != nil {
		if errors.Is(err, uplink.ErrObjectNotFound) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	return true, nil
}

func (o *ObjectStore) GetObject(bucket, key string) (io.ReadCloser, error) {
	o.log.Infof("objectStore.GetObject called")
	downloader, err := o.project.DownloadObject(context.Background(), bucket, key, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return downloader, nil
}

func (o *ObjectStore) ListCommonPrefixes(bucket, prefix, delimiter string) ([]string, error) {
	o.log.Infof("objectStore.ListCommonPrefixes called")
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

func (o *ObjectStore) ListObjects(bucket, prefix string) ([]string, error) {
	o.log.Infof("objectStore.ListObjects called")
	object := o.project.ListObjects(context.Background(), bucket, &uplink.ListObjectsOptions{Prefix: prefix})
	var res []string
	for object.Next() {
		res = append(res, object.Item().Key)
	}
	if err := object.Err(); err != nil {
		return res, err
	}
	return res, nil
}

func (o *ObjectStore) DeleteObject(bucket, key string) error {
	o.log.Infof("objectStore.DeleteObject called")
	if _, err := o.project.DeleteObject(context.Background(), bucket, key); err != nil {
		return err
	}
	return nil
}

func (o *ObjectStore) CreateSignedURL(bucket, key string, ttl time.Duration) (string, error) {
	o.log.Infof("objectStore.CreateSignedURL called")
	var sharePrefixes []uplink.SharePrefix
	sharePrefixes = append(sharePrefixes, uplink.SharePrefix{
		Bucket: bucket,
		Prefix: key,
	})

	permission := uplink.ReadOnlyPermission()
	permission.NotAfter = time.Now().Add(ttl)

	newAccess, err := o.access.Share(permission, sharePrefixes...)
	if err != nil {
		return "", err
	}

	newAccessData, err := newAccess.Serialize()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s/%s", linksharingBaseURL,
		url.PathEscape(newAccessData),
		url.PathEscape(bucket),
		url.PathEscape(key)), nil
}
