// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewStorageServiceTruncateByKeyspacePostParams creates a new StorageServiceTruncateByKeyspacePostParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewStorageServiceTruncateByKeyspacePostParams() *StorageServiceTruncateByKeyspacePostParams {
	return &StorageServiceTruncateByKeyspacePostParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewStorageServiceTruncateByKeyspacePostParamsWithTimeout creates a new StorageServiceTruncateByKeyspacePostParams object
// with the ability to set a timeout on a request.
func NewStorageServiceTruncateByKeyspacePostParamsWithTimeout(timeout time.Duration) *StorageServiceTruncateByKeyspacePostParams {
	return &StorageServiceTruncateByKeyspacePostParams{
		timeout: timeout,
	}
}

// NewStorageServiceTruncateByKeyspacePostParamsWithContext creates a new StorageServiceTruncateByKeyspacePostParams object
// with the ability to set a context for a request.
func NewStorageServiceTruncateByKeyspacePostParamsWithContext(ctx context.Context) *StorageServiceTruncateByKeyspacePostParams {
	return &StorageServiceTruncateByKeyspacePostParams{
		Context: ctx,
	}
}

// NewStorageServiceTruncateByKeyspacePostParamsWithHTTPClient creates a new StorageServiceTruncateByKeyspacePostParams object
// with the ability to set a custom HTTPClient for a request.
func NewStorageServiceTruncateByKeyspacePostParamsWithHTTPClient(client *http.Client) *StorageServiceTruncateByKeyspacePostParams {
	return &StorageServiceTruncateByKeyspacePostParams{
		HTTPClient: client,
	}
}

/*
StorageServiceTruncateByKeyspacePostParams contains all the parameters to send to the API endpoint

	for the storage service truncate by keyspace post operation.

	Typically these are written to a http.Request.
*/
type StorageServiceTruncateByKeyspacePostParams struct {

	/* Cf.

	   Column family name
	*/
	Cf *string

	/* Keyspace.

	   The keyspace
	*/
	Keyspace string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the storage service truncate by keyspace post params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *StorageServiceTruncateByKeyspacePostParams) WithDefaults() *StorageServiceTruncateByKeyspacePostParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the storage service truncate by keyspace post params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *StorageServiceTruncateByKeyspacePostParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the storage service truncate by keyspace post params
func (o *StorageServiceTruncateByKeyspacePostParams) WithTimeout(timeout time.Duration) *StorageServiceTruncateByKeyspacePostParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the storage service truncate by keyspace post params
func (o *StorageServiceTruncateByKeyspacePostParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the storage service truncate by keyspace post params
func (o *StorageServiceTruncateByKeyspacePostParams) WithContext(ctx context.Context) *StorageServiceTruncateByKeyspacePostParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the storage service truncate by keyspace post params
func (o *StorageServiceTruncateByKeyspacePostParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the storage service truncate by keyspace post params
func (o *StorageServiceTruncateByKeyspacePostParams) WithHTTPClient(client *http.Client) *StorageServiceTruncateByKeyspacePostParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the storage service truncate by keyspace post params
func (o *StorageServiceTruncateByKeyspacePostParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithCf adds the cf to the storage service truncate by keyspace post params
func (o *StorageServiceTruncateByKeyspacePostParams) WithCf(cf *string) *StorageServiceTruncateByKeyspacePostParams {
	o.SetCf(cf)
	return o
}

// SetCf adds the cf to the storage service truncate by keyspace post params
func (o *StorageServiceTruncateByKeyspacePostParams) SetCf(cf *string) {
	o.Cf = cf
}

// WithKeyspace adds the keyspace to the storage service truncate by keyspace post params
func (o *StorageServiceTruncateByKeyspacePostParams) WithKeyspace(keyspace string) *StorageServiceTruncateByKeyspacePostParams {
	o.SetKeyspace(keyspace)
	return o
}

// SetKeyspace adds the keyspace to the storage service truncate by keyspace post params
func (o *StorageServiceTruncateByKeyspacePostParams) SetKeyspace(keyspace string) {
	o.Keyspace = keyspace
}

// WriteToRequest writes these params to a swagger request
func (o *StorageServiceTruncateByKeyspacePostParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Cf != nil {

		// query param cf
		var qrCf string

		if o.Cf != nil {
			qrCf = *o.Cf
		}
		qCf := qrCf
		if qCf != "" {

			if err := r.SetQueryParam("cf", qCf); err != nil {
				return err
			}
		}
	}

	// path param keyspace
	if err := r.SetPathParam("keyspace", o.Keyspace); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
