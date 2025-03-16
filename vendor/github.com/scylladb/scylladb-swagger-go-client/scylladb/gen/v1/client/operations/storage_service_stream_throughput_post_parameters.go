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
	"github.com/go-openapi/swag"
)

// NewStorageServiceStreamThroughputPostParams creates a new StorageServiceStreamThroughputPostParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewStorageServiceStreamThroughputPostParams() *StorageServiceStreamThroughputPostParams {
	return &StorageServiceStreamThroughputPostParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewStorageServiceStreamThroughputPostParamsWithTimeout creates a new StorageServiceStreamThroughputPostParams object
// with the ability to set a timeout on a request.
func NewStorageServiceStreamThroughputPostParamsWithTimeout(timeout time.Duration) *StorageServiceStreamThroughputPostParams {
	return &StorageServiceStreamThroughputPostParams{
		timeout: timeout,
	}
}

// NewStorageServiceStreamThroughputPostParamsWithContext creates a new StorageServiceStreamThroughputPostParams object
// with the ability to set a context for a request.
func NewStorageServiceStreamThroughputPostParamsWithContext(ctx context.Context) *StorageServiceStreamThroughputPostParams {
	return &StorageServiceStreamThroughputPostParams{
		Context: ctx,
	}
}

// NewStorageServiceStreamThroughputPostParamsWithHTTPClient creates a new StorageServiceStreamThroughputPostParams object
// with the ability to set a custom HTTPClient for a request.
func NewStorageServiceStreamThroughputPostParamsWithHTTPClient(client *http.Client) *StorageServiceStreamThroughputPostParams {
	return &StorageServiceStreamThroughputPostParams{
		HTTPClient: client,
	}
}

/*
StorageServiceStreamThroughputPostParams contains all the parameters to send to the API endpoint

	for the storage service stream throughput post operation.

	Typically these are written to a http.Request.
*/
type StorageServiceStreamThroughputPostParams struct {

	/* Value.

	   Stream throughput

	   Format: int32
	*/
	Value int32

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the storage service stream throughput post params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *StorageServiceStreamThroughputPostParams) WithDefaults() *StorageServiceStreamThroughputPostParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the storage service stream throughput post params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *StorageServiceStreamThroughputPostParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the storage service stream throughput post params
func (o *StorageServiceStreamThroughputPostParams) WithTimeout(timeout time.Duration) *StorageServiceStreamThroughputPostParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the storage service stream throughput post params
func (o *StorageServiceStreamThroughputPostParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the storage service stream throughput post params
func (o *StorageServiceStreamThroughputPostParams) WithContext(ctx context.Context) *StorageServiceStreamThroughputPostParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the storage service stream throughput post params
func (o *StorageServiceStreamThroughputPostParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the storage service stream throughput post params
func (o *StorageServiceStreamThroughputPostParams) WithHTTPClient(client *http.Client) *StorageServiceStreamThroughputPostParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the storage service stream throughput post params
func (o *StorageServiceStreamThroughputPostParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithValue adds the value to the storage service stream throughput post params
func (o *StorageServiceStreamThroughputPostParams) WithValue(value int32) *StorageServiceStreamThroughputPostParams {
	o.SetValue(value)
	return o
}

// SetValue adds the value to the storage service stream throughput post params
func (o *StorageServiceStreamThroughputPostParams) SetValue(value int32) {
	o.Value = value
}

// WriteToRequest writes these params to a swagger request
func (o *StorageServiceStreamThroughputPostParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// query param value
	qrValue := o.Value
	qValue := swag.FormatInt32(qrValue)
	if qValue != "" {

		if err := r.SetQueryParam("value", qValue); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
