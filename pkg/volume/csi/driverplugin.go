package csi

//const (
//	csiTimeout = 2 * time.Minute
//)
//
//type DriverPlugin struct {
//	client csiClient
//}
//
//func NewDriverPlugin(driverName string, driverEndpoint string) (*DriverPlugin, error) {
//	klog.V(5).InfoS("NewDriverPlugin", "driverName", "driverEndpoint", driverEndpoint)
//	var err error
//
//	client := newDriverClient(driverName, driverEndpoint)
//
//	ctx, cancel := context.WithTimeout(context.Background(), csiTimeout)
//	defer cancel()
//
//	_, _, _, err = client.NodeGetInfo(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to call NodeGetInfo: %v", err)
//	}
//
//	return &DriverPlugin{
//		client: client,
//	}, nil
//}
