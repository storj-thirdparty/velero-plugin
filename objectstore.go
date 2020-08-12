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
	log    logrus.FieldLogger
	access *uplink.Access
}

func newObjectStore(logger logrus.FieldLogger) *ObjectStore {
	return &ObjectStore{log: logger}
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
	return nil
}

func (o *ObjectStore) PutObject(bucket, key string, body io.Reader) error {
	o.log.Infof("objectStore.PutObject called")
	project, err := uplink.OpenProject(context.Background(), o.access)
	if err != nil {
		return err
	}
	defer project.Close()
	upload, err := project.UploadObject(context.Background(), bucket, key, nil)
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
	project, err := uplink.OpenProject(context.Background(), o.access)
	if err != nil {
		return false, err
	}
	defer project.Close()

	if _, err := project.StatObject(context.Background(), bucket, key); err != nil {
		if errors.Is(err, uplink.ErrObjectNotFound) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	return true, nil
}

func (o *ObjectStore) GetObject(bucket, key string) (io.ReadCloser, error) {
	o.log.Infof("objectStore.GetObject called")
	project, err := uplink.OpenProject(context.Background(), o.access)
	if err != nil {
		return nil, err
	}
	defer project.Close()
	downloader, err := project.DownloadObject(context.Background(), bucket, key, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return downloader, nil
}

func (o *ObjectStore) ListCommonPrefixes(bucket, prefix, delimiter string) ([]string, error) {
	o.log.Infof("objectStore.ListCommonPrefixes called")
	project, err := uplink.OpenProject(context.Background(), o.access)
	if err != nil {
		return nil, err
	}
	defer project.Close()

	objectsIter := project.ListObjects(context.Background(), bucket, &uplink.ListObjectsOptions{Prefix: prefix})
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
	project, err := uplink.OpenProject(context.Background(), o.access)
	if err != nil {
		return nil, err
	}
	defer project.Close()

	object := project.ListObjects(context.Background(), bucket, &uplink.ListObjectsOptions{Prefix: prefix})
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
	project, err := uplink.OpenProject(context.Background(), o.access)
	if err != nil {
		return err
	}
	defer project.Close()

	if _, err := project.DeleteObject(context.Background(), bucket, key); err != nil {
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

	permission := uplink.Permission{}
	permission.AllowDownload = true
	permission.AllowDelete = false
	permission.AllowUpload = false
	permission.AllowList = true // ?
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
