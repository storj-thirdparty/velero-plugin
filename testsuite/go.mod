module storj.io/velero-plugin/testsuite

go 1.14

replace storj.io/velero-plugin => ../

require (
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.16.0
	storj.io/common v0.0.0-20210104180112-e8500e1c37a0
	storj.io/linksharing v1.5.0
	storj.io/storj v1.19.8
	storj.io/velero-plugin v0.0.0-00010101000000-000000000000
)
