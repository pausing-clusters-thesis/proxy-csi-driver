package resourceapply

import (
	"context"

	"github.com/scylladb/scylla-operator/pkg/resourceapply"
	storagev1 "k8s.io/api/storage/v1"
	storagev1client "k8s.io/client-go/kubernetes/typed/storage/v1"
	storagev1listers "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/record"
)

func ApplyVolumeAttachmentWithControl(
	ctx context.Context,
	control resourceapply.ApplyControlInterface[*storagev1.VolumeAttachment],
	recorder record.EventRecorder,
	required *storagev1.VolumeAttachment,
	options resourceapply.ApplyOptions,
) (*storagev1.VolumeAttachment, bool, error) {
	return resourceapply.ApplyGeneric[*storagev1.VolumeAttachment](ctx, control, recorder, required, options)
}

func ApplyVolumeAttachment(
	ctx context.Context,
	client storagev1client.VolumeAttachmentsGetter,
	lister storagev1listers.VolumeAttachmentLister,
	recorder record.EventRecorder,
	required *storagev1.VolumeAttachment,
	options resourceapply.ApplyOptions,
) (*storagev1.VolumeAttachment, bool, error) {
	return ApplyVolumeAttachmentWithControl(
		ctx,
		resourceapply.ApplyControlFuncs[*storagev1.VolumeAttachment]{
			GetCachedFunc: lister.Get,
			CreateFunc:    client.VolumeAttachments().Create,
			UpdateFunc:    client.VolumeAttachments().Update,
			DeleteFunc:    client.VolumeAttachments().Delete,
		},
		recorder,
		required,
		options,
	)
}
