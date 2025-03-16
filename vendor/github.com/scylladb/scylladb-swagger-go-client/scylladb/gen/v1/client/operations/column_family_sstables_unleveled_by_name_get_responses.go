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

// ColumnFamilySstablesUnleveledByNameGetReader is a Reader for the ColumnFamilySstablesUnleveledByNameGet structure.
type ColumnFamilySstablesUnleveledByNameGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ColumnFamilySstablesUnleveledByNameGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewColumnFamilySstablesUnleveledByNameGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewColumnFamilySstablesUnleveledByNameGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewColumnFamilySstablesUnleveledByNameGetOK creates a ColumnFamilySstablesUnleveledByNameGetOK with default headers values
func NewColumnFamilySstablesUnleveledByNameGetOK() *ColumnFamilySstablesUnleveledByNameGetOK {
	return &ColumnFamilySstablesUnleveledByNameGetOK{}
}

/*
ColumnFamilySstablesUnleveledByNameGetOK handles this case with default header values.

ColumnFamilySstablesUnleveledByNameGetOK column family sstables unleveled by name get o k
*/
type ColumnFamilySstablesUnleveledByNameGetOK struct {
	Payload []string
}

func (o *ColumnFamilySstablesUnleveledByNameGetOK) GetPayload() []string {
	return o.Payload
}

func (o *ColumnFamilySstablesUnleveledByNameGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewColumnFamilySstablesUnleveledByNameGetDefault creates a ColumnFamilySstablesUnleveledByNameGetDefault with default headers values
func NewColumnFamilySstablesUnleveledByNameGetDefault(code int) *ColumnFamilySstablesUnleveledByNameGetDefault {
	return &ColumnFamilySstablesUnleveledByNameGetDefault{
		_statusCode: code,
	}
}

/*
ColumnFamilySstablesUnleveledByNameGetDefault handles this case with default header values.

internal server error
*/
type ColumnFamilySstablesUnleveledByNameGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the column family sstables unleveled by name get default response
func (o *ColumnFamilySstablesUnleveledByNameGetDefault) Code() int {
	return o._statusCode
}

func (o *ColumnFamilySstablesUnleveledByNameGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *ColumnFamilySstablesUnleveledByNameGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *ColumnFamilySstablesUnleveledByNameGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
