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

// NewStorageServiceRepairAsyncByKeyspaceGetParams creates a new StorageServiceRepairAsyncByKeyspaceGetParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewStorageServiceRepairAsyncByKeyspaceGetParams() *StorageServiceRepairAsyncByKeyspaceGetParams {
	return &StorageServiceRepairAsyncByKeyspaceGetParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewStorageServiceRepairAsyncByKeyspaceGetParamsWithTimeout creates a new StorageServiceRepairAsyncByKeyspaceGetParams object
// with the ability to set a timeout on a request.
func NewStorageServiceRepairAsyncByKeyspaceGetParamsWithTimeout(timeout time.Duration) *StorageServiceRepairAsyncByKeyspaceGetParams {
	return &StorageServiceRepairAsyncByKeyspaceGetParams{
		timeout: timeout,
	}
}

// NewStorageServiceRepairAsyncByKeyspaceGetParamsWithContext creates a new StorageServiceRepairAsyncByKeyspaceGetParams object
// with the ability to set a context for a request.
func NewStorageServiceRepairAsyncByKeyspaceGetParamsWithContext(ctx context.Context) *StorageServiceRepairAsyncByKeyspaceGetParams {
	return &StorageServiceRepairAsyncByKeyspaceGetParams{
		Context: ctx,
	}
}

// NewStorageServiceRepairAsyncByKeyspaceGetParamsWithHTTPClient creates a new StorageServiceRepairAsyncByKeyspaceGetParams object
// with the ability to set a custom HTTPClient for a request.
func NewStorageServiceRepairAsyncByKeyspaceGetParamsWithHTTPClient(client *http.Client) *StorageServiceRepairAsyncByKeyspaceGetParams {
	return &StorageServiceRepairAsyncByKeyspaceGetParams{
		HTTPClient: client,
	}
}

/*
StorageServiceRepairAsyncByKeyspaceGetParams contains all the parameters to send to the API endpoint

	for the storage service repair async by keyspace get operation.

	Typically these are written to a http.Request.
*/
type StorageServiceRepairAsyncByKeyspaceGetParams struct {

	/* ID.

	   The repair ID to check for status

	   Format: int32
	*/
	ID int32

	/* Keyspace.

	   The keyspace repair is running on
	*/
	Keyspace string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the storage service repair async by keyspace get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) WithDefaults() *StorageServiceRepairAsyncByKeyspaceGetParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the storage service repair async by keyspace get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the storage service repair async by keyspace get params
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) WithTimeout(timeout time.Duration) *StorageServiceRepairAsyncByKeyspaceGetParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the storage service repair async by keyspace get params
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the storage service repair async by keyspace get params
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) WithContext(ctx context.Context) *StorageServiceRepairAsyncByKeyspaceGetParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the storage service repair async by keyspace get params
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the storage service repair async by keyspace get params
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) WithHTTPClient(client *http.Client) *StorageServiceRepairAsyncByKeyspaceGetParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the storage service repair async by keyspace get params
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithID adds the id to the storage service repair async by keyspace get params
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) WithID(id int32) *StorageServiceRepairAsyncByKeyspaceGetParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the storage service repair async by keyspace get params
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) SetID(id int32) {
	o.ID = id
}

// WithKeyspace adds the keyspace to the storage service repair async by keyspace get params
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) WithKeyspace(keyspace string) *StorageServiceRepairAsyncByKeyspaceGetParams {
	o.SetKeyspace(keyspace)
	return o
}

// SetKeyspace adds the keyspace to the storage service repair async by keyspace get params
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) SetKeyspace(keyspace string) {
	o.Keyspace = keyspace
}

// WriteToRequest writes these params to a swagger request
func (o *StorageServiceRepairAsyncByKeyspaceGetParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// query param id
	qrID := o.ID
	qID := swag.FormatInt32(qrID)
	if qID != "" {

		if err := r.SetQueryParam("id", qID); err != nil {
			return err
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
