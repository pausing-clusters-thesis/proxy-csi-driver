package csi

import (
	"context"
	"fmt"

	"github.com/kubernetes-csi/csi-lib-utils/connection"
	"github.com/kubernetes-csi/csi-lib-utils/rpc"
	"k8s.io/klog/v2"
)

type DriverRegistry struct {
	registry map[string]string
}

func NewDriverRegistry() *DriverRegistry {
	return &DriverRegistry{
		registry: make(map[string]string),
	}
}

// TODO: registry shouldn't fail on an unsuccessful connection, it should run a prober in the background and add/remove drivers from registry as they become available
// TODO: in the current state the entire driver will fail if any of the registered drivers is not available on startup, and it won't update in runtime
// TODO: also look into plugging into registry path used by kubelet, maybe we could use those events in runtime
func (r *DriverRegistry) RegisterDriver(ctx context.Context, address string) error {
	klog.V(5).InfoS("DriverRegistry.RegisterDriver", "address", address)

	klog.V(5).InfoS("DriverRegistry.RegisterDriver: attempting to open a gRPC connection", "address", address)
	conn, err := connection.Connect(ctx, address, nil)
	if err != nil {
		return fmt.Errorf("can't connect to socket %q: %w", address, err)
	}
	defer conn.Close()

	driverName, err := rpc.GetDriverName(ctx, conn)
	if err != nil {
		return fmt.Errorf("can't get driver name: %w", err)
	}

	r.registry[driverName] = address

	return nil
}

func (r *DriverRegistry) GetAddress(driverName string) (string, error) {
	klog.V(5).InfoS("DriverRegistry.GetAddress")

	address, ok := r.registry[driverName]
	if !ok {
		return "", fmt.Errorf("driver %q not in registry", driverName)
	}

	return address, nil
}
