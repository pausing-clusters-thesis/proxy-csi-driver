package scheme

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	Scheme                = runtime.NewScheme()
	Codecs                = serializer.NewCodecFactory(Scheme, serializer.EnableStrict)
	DefaultYamlSerializer = json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		Scheme,
		Scheme,
		json.SerializerOptions{
			Yaml:   true,
			Pretty: false,
			Strict: true,
		},
	)
)

func init() {
	utilruntime.Must(kscheme.AddToScheme(Scheme))
}
