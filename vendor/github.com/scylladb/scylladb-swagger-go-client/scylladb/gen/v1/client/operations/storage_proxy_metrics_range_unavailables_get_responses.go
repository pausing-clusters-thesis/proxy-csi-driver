// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/scylladb/scylladb-swagger-go-client/scylladb/gen/v1/models"
)

// StorageProxyMetricsRangeUnavailablesGetReader is a Reader for the StorageProxyMetricsRangeUnavailablesGet structure.
type StorageProxyMetricsRangeUnavailablesGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *StorageProxyMetricsRangeUnavailablesGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewStorageProxyMetricsRangeUnavailablesGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewStorageProxyMetricsRangeUnavailablesGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewStorageProxyMetricsRangeUnavailablesGetOK creates a StorageProxyMetricsRangeUnavailablesGetOK with default headers values
func NewStorageProxyMetricsRangeUnavailablesGetOK() *StorageProxyMetricsRangeUnavailablesGetOK {
	return &StorageProxyMetricsRangeUnavailablesGetOK{}
}

/*
StorageProxyMetricsRangeUnavailablesGetOK handles this case with default header values.

StorageProxyMetricsRangeUnavailablesGetOK storage proxy metrics range unavailables get o k
*/
type StorageProxyMetricsRangeUnavailablesGetOK struct {
	Payload int32
}

func (o *StorageProxyMetricsRangeUnavailablesGetOK) GetPayload() int32 {
	return o.Payload
}

func (o *StorageProxyMetricsRangeUnavailablesGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewStorageProxyMetricsRangeUnavailablesGetDefault creates a StorageProxyMetricsRangeUnavailablesGetDefault with default headers values
func NewStorageProxyMetricsRangeUnavailablesGetDefault(code int) *StorageProxyMetricsRangeUnavailablesGetDefault {
	return &StorageProxyMetricsRangeUnavailablesGetDefault{
		_statusCode: code,
	}
}

/*
StorageProxyMetricsRangeUnavailablesGetDefault handles this case with default header values.

internal server error
*/
type StorageProxyMetricsRangeUnavailablesGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the storage proxy metrics range unavailables get default response
func (o *StorageProxyMetricsRangeUnavailablesGetDefault) Code() int {
	return o._statusCode
}

func (o *StorageProxyMetricsRangeUnavailablesGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *StorageProxyMetricsRangeUnavailablesGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *StorageProxyMetricsRangeUnavailablesGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
