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

// NewColumnFamilySstablesPerLevelByNameGetParams creates a new ColumnFamilySstablesPerLevelByNameGetParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewColumnFamilySstablesPerLevelByNameGetParams() *ColumnFamilySstablesPerLevelByNameGetParams {
	return &ColumnFamilySstablesPerLevelByNameGetParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewColumnFamilySstablesPerLevelByNameGetParamsWithTimeout creates a new ColumnFamilySstablesPerLevelByNameGetParams object
// with the ability to set a timeout on a request.
func NewColumnFamilySstablesPerLevelByNameGetParamsWithTimeout(timeout time.Duration) *ColumnFamilySstablesPerLevelByNameGetParams {
	return &ColumnFamilySstablesPerLevelByNameGetParams{
		timeout: timeout,
	}
}

// NewColumnFamilySstablesPerLevelByNameGetParamsWithContext creates a new ColumnFamilySstablesPerLevelByNameGetParams object
// with the ability to set a context for a request.
func NewColumnFamilySstablesPerLevelByNameGetParamsWithContext(ctx context.Context) *ColumnFamilySstablesPerLevelByNameGetParams {
	return &ColumnFamilySstablesPerLevelByNameGetParams{
		Context: ctx,
	}
}

// NewColumnFamilySstablesPerLevelByNameGetParamsWithHTTPClient creates a new ColumnFamilySstablesPerLevelByNameGetParams object
// with the ability to set a custom HTTPClient for a request.
func NewColumnFamilySstablesPerLevelByNameGetParamsWithHTTPClient(client *http.Client) *ColumnFamilySstablesPerLevelByNameGetParams {
	return &ColumnFamilySstablesPerLevelByNameGetParams{
		HTTPClient: client,
	}
}

/*
ColumnFamilySstablesPerLevelByNameGetParams contains all the parameters to send to the API endpoint

	for the column family sstables per level by name get operation.

	Typically these are written to a http.Request.
*/
type ColumnFamilySstablesPerLevelByNameGetParams struct {

	/* Name.

	   The column family name in keyspace:name format
	*/
	Name string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the column family sstables per level by name get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ColumnFamilySstablesPerLevelByNameGetParams) WithDefaults() *ColumnFamilySstablesPerLevelByNameGetParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the column family sstables per level by name get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ColumnFamilySstablesPerLevelByNameGetParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the column family sstables per level by name get params
func (o *ColumnFamilySstablesPerLevelByNameGetParams) WithTimeout(timeout time.Duration) *ColumnFamilySstablesPerLevelByNameGetParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the column family sstables per level by name get params
func (o *ColumnFamilySstablesPerLevelByNameGetParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the column family sstables per level by name get params
func (o *ColumnFamilySstablesPerLevelByNameGetParams) WithContext(ctx context.Context) *ColumnFamilySstablesPerLevelByNameGetParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the column family sstables per level by name get params
func (o *ColumnFamilySstablesPerLevelByNameGetParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the column family sstables per level by name get params
func (o *ColumnFamilySstablesPerLevelByNameGetParams) WithHTTPClient(client *http.Client) *ColumnFamilySstablesPerLevelByNameGetParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the column family sstables per level by name get params
func (o *ColumnFamilySstablesPerLevelByNameGetParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithName adds the name to the column family sstables per level by name get params
func (o *ColumnFamilySstablesPerLevelByNameGetParams) WithName(name string) *ColumnFamilySstablesPerLevelByNameGetParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the column family sstables per level by name get params
func (o *ColumnFamilySstablesPerLevelByNameGetParams) SetName(name string) {
	o.Name = name
}

// WriteToRequest writes these params to a swagger request
func (o *ColumnFamilySstablesPerLevelByNameGetParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param name
	if err := r.SetPathParam("name", o.Name); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
