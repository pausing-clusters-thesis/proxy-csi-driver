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

// NewColumnFamilyMetricsWriteLatencyHistogramByNameGetParams creates a new ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewColumnFamilyMetricsWriteLatencyHistogramByNameGetParams() *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams {
	return &ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewColumnFamilyMetricsWriteLatencyHistogramByNameGetParamsWithTimeout creates a new ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams object
// with the ability to set a timeout on a request.
func NewColumnFamilyMetricsWriteLatencyHistogramByNameGetParamsWithTimeout(timeout time.Duration) *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams {
	return &ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams{
		timeout: timeout,
	}
}

// NewColumnFamilyMetricsWriteLatencyHistogramByNameGetParamsWithContext creates a new ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams object
// with the ability to set a context for a request.
func NewColumnFamilyMetricsWriteLatencyHistogramByNameGetParamsWithContext(ctx context.Context) *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams {
	return &ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams{
		Context: ctx,
	}
}

// NewColumnFamilyMetricsWriteLatencyHistogramByNameGetParamsWithHTTPClient creates a new ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams object
// with the ability to set a custom HTTPClient for a request.
func NewColumnFamilyMetricsWriteLatencyHistogramByNameGetParamsWithHTTPClient(client *http.Client) *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams {
	return &ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams{
		HTTPClient: client,
	}
}

/*
ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams contains all the parameters to send to the API endpoint

	for the column family metrics write latency histogram by name get operation.

	Typically these are written to a http.Request.
*/
type ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams struct {

	/* Name.

	   The column family name in keyspace:name format
	*/
	Name string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the column family metrics write latency histogram by name get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams) WithDefaults() *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the column family metrics write latency histogram by name get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the column family metrics write latency histogram by name get params
func (o *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams) WithTimeout(timeout time.Duration) *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the column family metrics write latency histogram by name get params
func (o *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the column family metrics write latency histogram by name get params
func (o *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams) WithContext(ctx context.Context) *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the column family metrics write latency histogram by name get params
func (o *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the column family metrics write latency histogram by name get params
func (o *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams) WithHTTPClient(client *http.Client) *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the column family metrics write latency histogram by name get params
func (o *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithName adds the name to the column family metrics write latency histogram by name get params
func (o *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams) WithName(name string) *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the column family metrics write latency histogram by name get params
func (o *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams) SetName(name string) {
	o.Name = name
}

// WriteToRequest writes these params to a swagger request
func (o *ColumnFamilyMetricsWriteLatencyHistogramByNameGetParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
