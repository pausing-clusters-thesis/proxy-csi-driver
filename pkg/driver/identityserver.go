package driver

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"github.com/pausing-clusters-thesis/proxy-csi-driver/pkg/naming"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/klog/v2"
)

type IdentityServer struct {
	csi.UnimplementedIdentityServer

	version string
}

var _ csi.IdentityServer = &IdentityServer{}

func (is *IdentityServer) GetPluginInfo(ctx context.Context, request *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	klog.V(5).InfoS("Request received", "server", "identity", "function", "GetPluginInfo", "request", protosanitizer.StripSecrets(request))

	return &csi.GetPluginInfoResponse{
		Name:          naming.DriverName,
		VendorVersion: is.version,
		Manifest:      nil,
	}, nil
}

func (is *IdentityServer) GetPluginCapabilities(ctx context.Context, request *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	klog.V(5).InfoS("Request received", "server", "identity", "function", "GetPluginCapabilities", "request", protosanitizer.StripSecrets(request))

	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}

func (is *IdentityServer) Probe(ctx context.Context, request *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	klog.V(5).InfoS("Request received", "server", "identity", "function", "Probe", "request", protosanitizer.StripSecrets(request))

	return &csi.ProbeResponse{
		Ready: wrapperspb.Bool(true),
	}, nil
}
