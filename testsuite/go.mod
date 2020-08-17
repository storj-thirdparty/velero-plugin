module storj.io/velero-plugin/testsuite

go 1.14

replace storj.io/velero-plugin => ../

require (
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.5.1
	go.uber.org/zap v1.15.0
	storj.io/common v0.0.0-20200729140050-4c1ddac6fa63
	storj.io/linksharing v1.0.0
	storj.io/storj v0.12.1-0.20200803142802-935f44ddb7da
	storj.io/velero-plugin v0.0.0-00010101000000-000000000000
)
