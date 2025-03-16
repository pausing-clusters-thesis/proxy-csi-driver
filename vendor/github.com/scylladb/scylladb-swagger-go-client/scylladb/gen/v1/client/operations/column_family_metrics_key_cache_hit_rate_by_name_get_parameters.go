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

// NewColumnFamilyMetricsKeyCacheHitRateByNameGetParams creates a new ColumnFamilyMetricsKeyCacheHitRateByNameGetParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewColumnFamilyMetricsKeyCacheHitRateByNameGetParams() *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams {
	return &ColumnFamilyMetricsKeyCacheHitRateByNameGetParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewColumnFamilyMetricsKeyCacheHitRateByNameGetParamsWithTimeout creates a new ColumnFamilyMetricsKeyCacheHitRateByNameGetParams object
// with the ability to set a timeout on a request.
func NewColumnFamilyMetricsKeyCacheHitRateByNameGetParamsWithTimeout(timeout time.Duration) *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams {
	return &ColumnFamilyMetricsKeyCacheHitRateByNameGetParams{
		timeout: timeout,
	}
}

// NewColumnFamilyMetricsKeyCacheHitRateByNameGetParamsWithContext creates a new ColumnFamilyMetricsKeyCacheHitRateByNameGetParams object
// with the ability to set a context for a request.
func NewColumnFamilyMetricsKeyCacheHitRateByNameGetParamsWithContext(ctx context.Context) *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams {
	return &ColumnFamilyMetricsKeyCacheHitRateByNameGetParams{
		Context: ctx,
	}
}

// NewColumnFamilyMetricsKeyCacheHitRateByNameGetParamsWithHTTPClient creates a new ColumnFamilyMetricsKeyCacheHitRateByNameGetParams object
// with the ability to set a custom HTTPClient for a request.
func NewColumnFamilyMetricsKeyCacheHitRateByNameGetParamsWithHTTPClient(client *http.Client) *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams {
	return &ColumnFamilyMetricsKeyCacheHitRateByNameGetParams{
		HTTPClient: client,
	}
}

/*
ColumnFamilyMetricsKeyCacheHitRateByNameGetParams contains all the parameters to send to the API endpoint

	for the column family metrics key cache hit rate by name get operation.

	Typically these are written to a http.Request.
*/
type ColumnFamilyMetricsKeyCacheHitRateByNameGetParams struct {

	/* Name.

	   The column family name in keyspace:name format
	*/
	Name string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the column family metrics key cache hit rate by name get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams) WithDefaults() *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the column family metrics key cache hit rate by name get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the column family metrics key cache hit rate by name get params
func (o *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams) WithTimeout(timeout time.Duration) *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the column family metrics key cache hit rate by name get params
func (o *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the column family metrics key cache hit rate by name get params
func (o *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams) WithContext(ctx context.Context) *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the column family metrics key cache hit rate by name get params
func (o *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the column family metrics key cache hit rate by name get params
func (o *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams) WithHTTPClient(client *http.Client) *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the column family metrics key cache hit rate by name get params
func (o *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithName adds the name to the column family metrics key cache hit rate by name get params
func (o *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams) WithName(name string) *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the column family metrics key cache hit rate by name get params
func (o *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams) SetName(name string) {
	o.Name = name
}

// WriteToRequest writes these params to a swagger request
func (o *ColumnFamilyMetricsKeyCacheHitRateByNameGetParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
