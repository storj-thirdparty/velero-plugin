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

// config params
const (
	accessGrant        = "accessGrant"
	linksharingBaseURL = "linksharingBaseURL"
)

const defaultLinksharingBaseURL = "https://link.tardigradeshare.io"

type ObjectStore struct {
	log                logrus.FieldLogger
	access             *uplink.Access
	project            *uplink.Project
	linksharingBaseURL string
}

func NewObjectStore(logger logrus.FieldLogger) *ObjectStore {
	return &ObjectStore{log: logger}
}

func (o *ObjectStore) Init(config map[string]string) error {
	o.log.Infof("objectStore.Init called")
	err := veleroplugin.ValidateObjectStoreConfigKeys(config, accessGrant, linksharingBaseURL)
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

	o.linksharingBaseURL = config[linksharingBaseURL]
	if o.linksharingBaseURL == "" {
		o.linksharingBaseURL = defaultLinksharingBaseURL
	}

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

	return fmt.Sprintf("%s/%s/%s/%s", o.linksharingBaseURL,
		url.PathEscape(restrictedAccessGrant),
		url.PathEscape(bucket),
		url.PathEscape(key)), nil
}
