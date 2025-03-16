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

// NewCommitlogMetricsTotalCommitLogSizeGetParams creates a new CommitlogMetricsTotalCommitLogSizeGetParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewCommitlogMetricsTotalCommitLogSizeGetParams() *CommitlogMetricsTotalCommitLogSizeGetParams {
	return &CommitlogMetricsTotalCommitLogSizeGetParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewCommitlogMetricsTotalCommitLogSizeGetParamsWithTimeout creates a new CommitlogMetricsTotalCommitLogSizeGetParams object
// with the ability to set a timeout on a request.
func NewCommitlogMetricsTotalCommitLogSizeGetParamsWithTimeout(timeout time.Duration) *CommitlogMetricsTotalCommitLogSizeGetParams {
	return &CommitlogMetricsTotalCommitLogSizeGetParams{
		timeout: timeout,
	}
}

// NewCommitlogMetricsTotalCommitLogSizeGetParamsWithContext creates a new CommitlogMetricsTotalCommitLogSizeGetParams object
// with the ability to set a context for a request.
func NewCommitlogMetricsTotalCommitLogSizeGetParamsWithContext(ctx context.Context) *CommitlogMetricsTotalCommitLogSizeGetParams {
	return &CommitlogMetricsTotalCommitLogSizeGetParams{
		Context: ctx,
	}
}

// NewCommitlogMetricsTotalCommitLogSizeGetParamsWithHTTPClient creates a new CommitlogMetricsTotalCommitLogSizeGetParams object
// with the ability to set a custom HTTPClient for a request.
func NewCommitlogMetricsTotalCommitLogSizeGetParamsWithHTTPClient(client *http.Client) *CommitlogMetricsTotalCommitLogSizeGetParams {
	return &CommitlogMetricsTotalCommitLogSizeGetParams{
		HTTPClient: client,
	}
}

/*
CommitlogMetricsTotalCommitLogSizeGetParams contains all the parameters to send to the API endpoint

	for the commitlog metrics total commit log size get operation.

	Typically these are written to a http.Request.
*/
type CommitlogMetricsTotalCommitLogSizeGetParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the commitlog metrics total commit log size get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *CommitlogMetricsTotalCommitLogSizeGetParams) WithDefaults() *CommitlogMetricsTotalCommitLogSizeGetParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the commitlog metrics total commit log size get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *CommitlogMetricsTotalCommitLogSizeGetParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the commitlog metrics total commit log size get params
func (o *CommitlogMetricsTotalCommitLogSizeGetParams) WithTimeout(timeout time.Duration) *CommitlogMetricsTotalCommitLogSizeGetParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the commitlog metrics total commit log size get params
func (o *CommitlogMetricsTotalCommitLogSizeGetParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the commitlog metrics total commit log size get params
func (o *CommitlogMetricsTotalCommitLogSizeGetParams) WithContext(ctx context.Context) *CommitlogMetricsTotalCommitLogSizeGetParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the commitlog metrics total commit log size get params
func (o *CommitlogMetricsTotalCommitLogSizeGetParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the commitlog metrics total commit log size get params
func (o *CommitlogMetricsTotalCommitLogSizeGetParams) WithHTTPClient(client *http.Client) *CommitlogMetricsTotalCommitLogSizeGetParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the commitlog metrics total commit log size get params
func (o *CommitlogMetricsTotalCommitLogSizeGetParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *CommitlogMetricsTotalCommitLogSizeGetParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
