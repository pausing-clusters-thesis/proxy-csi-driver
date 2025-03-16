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

// ColumnFamilyMetricsSstablesPerReadHistogramByNameGetReader is a Reader for the ColumnFamilyMetricsSstablesPerReadHistogramByNameGet structure.
type ColumnFamilyMetricsSstablesPerReadHistogramByNameGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ColumnFamilyMetricsSstablesPerReadHistogramByNameGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewColumnFamilyMetricsSstablesPerReadHistogramByNameGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewColumnFamilyMetricsSstablesPerReadHistogramByNameGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewColumnFamilyMetricsSstablesPerReadHistogramByNameGetOK creates a ColumnFamilyMetricsSstablesPerReadHistogramByNameGetOK with default headers values
func NewColumnFamilyMetricsSstablesPerReadHistogramByNameGetOK() *ColumnFamilyMetricsSstablesPerReadHistogramByNameGetOK {
	return &ColumnFamilyMetricsSstablesPerReadHistogramByNameGetOK{}
}

/*
ColumnFamilyMetricsSstablesPerReadHistogramByNameGetOK handles this case with default header values.

ColumnFamilyMetricsSstablesPerReadHistogramByNameGetOK column family metrics sstables per read histogram by name get o k
*/
type ColumnFamilyMetricsSstablesPerReadHistogramByNameGetOK struct {
	Payload interface{}
}

func (o *ColumnFamilyMetricsSstablesPerReadHistogramByNameGetOK) GetPayload() interface{} {
	return o.Payload
}

func (o *ColumnFamilyMetricsSstablesPerReadHistogramByNameGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewColumnFamilyMetricsSstablesPerReadHistogramByNameGetDefault creates a ColumnFamilyMetricsSstablesPerReadHistogramByNameGetDefault with default headers values
func NewColumnFamilyMetricsSstablesPerReadHistogramByNameGetDefault(code int) *ColumnFamilyMetricsSstablesPerReadHistogramByNameGetDefault {
	return &ColumnFamilyMetricsSstablesPerReadHistogramByNameGetDefault{
		_statusCode: code,
	}
}

/*
ColumnFamilyMetricsSstablesPerReadHistogramByNameGetDefault handles this case with default header values.

internal server error
*/
type ColumnFamilyMetricsSstablesPerReadHistogramByNameGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the column family metrics sstables per read histogram by name get default response
func (o *ColumnFamilyMetricsSstablesPerReadHistogramByNameGetDefault) Code() int {
	return o._statusCode
}

func (o *ColumnFamilyMetricsSstablesPerReadHistogramByNameGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *ColumnFamilyMetricsSstablesPerReadHistogramByNameGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *ColumnFamilyMetricsSstablesPerReadHistogramByNameGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
