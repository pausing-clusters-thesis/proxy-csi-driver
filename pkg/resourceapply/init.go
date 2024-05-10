package resourceapply

import (
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/scheme"
	"github.com/scylladb/scylla-operator/pkg/resource"
)

func init() {
	resource.Scheme = scheme.Scheme
}
