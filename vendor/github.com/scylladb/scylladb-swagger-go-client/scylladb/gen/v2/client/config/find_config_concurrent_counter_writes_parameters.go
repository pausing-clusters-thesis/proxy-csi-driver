// Code generated by go-swagger; DO NOT EDIT.

package config

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

// NewFindConfigConcurrentCounterWritesParams creates a new FindConfigConcurrentCounterWritesParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewFindConfigConcurrentCounterWritesParams() *FindConfigConcurrentCounterWritesParams {
	return &FindConfigConcurrentCounterWritesParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewFindConfigConcurrentCounterWritesParamsWithTimeout creates a new FindConfigConcurrentCounterWritesParams object
// with the ability to set a timeout on a request.
func NewFindConfigConcurrentCounterWritesParamsWithTimeout(timeout time.Duration) *FindConfigConcurrentCounterWritesParams {
	return &FindConfigConcurrentCounterWritesParams{
		timeout: timeout,
	}
}

// NewFindConfigConcurrentCounterWritesParamsWithContext creates a new FindConfigConcurrentCounterWritesParams object
// with the ability to set a context for a request.
func NewFindConfigConcurrentCounterWritesParamsWithContext(ctx context.Context) *FindConfigConcurrentCounterWritesParams {
	return &FindConfigConcurrentCounterWritesParams{
		Context: ctx,
	}
}

// NewFindConfigConcurrentCounterWritesParamsWithHTTPClient creates a new FindConfigConcurrentCounterWritesParams object
// with the ability to set a custom HTTPClient for a request.
func NewFindConfigConcurrentCounterWritesParamsWithHTTPClient(client *http.Client) *FindConfigConcurrentCounterWritesParams {
	return &FindConfigConcurrentCounterWritesParams{
		HTTPClient: client,
	}
}

/*
FindConfigConcurrentCounterWritesParams contains all the parameters to send to the API endpoint

	for the find config concurrent counter writes operation.

	Typically these are written to a http.Request.
*/
type FindConfigConcurrentCounterWritesParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the find config concurrent counter writes params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *FindConfigConcurrentCounterWritesParams) WithDefaults() *FindConfigConcurrentCounterWritesParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the find config concurrent counter writes params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *FindConfigConcurrentCounterWritesParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the find config concurrent counter writes params
func (o *FindConfigConcurrentCounterWritesParams) WithTimeout(timeout time.Duration) *FindConfigConcurrentCounterWritesParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the find config concurrent counter writes params
func (o *FindConfigConcurrentCounterWritesParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the find config concurrent counter writes params
func (o *FindConfigConcurrentCounterWritesParams) WithContext(ctx context.Context) *FindConfigConcurrentCounterWritesParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the find config concurrent counter writes params
func (o *FindConfigConcurrentCounterWritesParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the find config concurrent counter writes params
func (o *FindConfigConcurrentCounterWritesParams) WithHTTPClient(client *http.Client) *FindConfigConcurrentCounterWritesParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the find config concurrent counter writes params
func (o *FindConfigConcurrentCounterWritesParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *FindConfigConcurrentCounterWritesParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
