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

// StorageProxyMetricsWriteHistogramGetReader is a Reader for the StorageProxyMetricsWriteHistogramGet structure.
type StorageProxyMetricsWriteHistogramGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *StorageProxyMetricsWriteHistogramGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewStorageProxyMetricsWriteHistogramGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewStorageProxyMetricsWriteHistogramGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewStorageProxyMetricsWriteHistogramGetOK creates a StorageProxyMetricsWriteHistogramGetOK with default headers values
func NewStorageProxyMetricsWriteHistogramGetOK() *StorageProxyMetricsWriteHistogramGetOK {
	return &StorageProxyMetricsWriteHistogramGetOK{}
}

/*
StorageProxyMetricsWriteHistogramGetOK handles this case with default header values.

StorageProxyMetricsWriteHistogramGetOK storage proxy metrics write histogram get o k
*/
type StorageProxyMetricsWriteHistogramGetOK struct {
}

func (o *StorageProxyMetricsWriteHistogramGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewStorageProxyMetricsWriteHistogramGetDefault creates a StorageProxyMetricsWriteHistogramGetDefault with default headers values
func NewStorageProxyMetricsWriteHistogramGetDefault(code int) *StorageProxyMetricsWriteHistogramGetDefault {
	return &StorageProxyMetricsWriteHistogramGetDefault{
		_statusCode: code,
	}
}

/*
StorageProxyMetricsWriteHistogramGetDefault handles this case with default header values.

internal server error
*/
type StorageProxyMetricsWriteHistogramGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the storage proxy metrics write histogram get default response
func (o *StorageProxyMetricsWriteHistogramGetDefault) Code() int {
	return o._statusCode
}

func (o *StorageProxyMetricsWriteHistogramGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *StorageProxyMetricsWriteHistogramGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *StorageProxyMetricsWriteHistogramGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
