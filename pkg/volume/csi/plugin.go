package csi

type Driver struct {
	endpoint string
}

type driverStore map[string]*Driver

type Plugin struct {
	driverStore
}

// Ideas:
/*
- have a csi plugin, which can handle multiple drivers
- drivers are registered against the csi plugin
- 	OR a separate csi driver store, which could then be extended with a watcher etc
- do we need a separate mounter? device mounter? attacher? couldn't all this be a part of the plugin?
- plugin needs kube client, csi client etc
-

*/
