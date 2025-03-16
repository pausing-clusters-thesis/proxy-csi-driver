// Code generated by go-swagger; DO NOT EDIT.

package config

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/scylladb/scylladb-swagger-go-client/scylladb/gen/v2/models"
)

// FindConfigPartitionerReader is a Reader for the FindConfigPartitioner structure.
type FindConfigPartitionerReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *FindConfigPartitionerReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewFindConfigPartitionerOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewFindConfigPartitionerDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewFindConfigPartitionerOK creates a FindConfigPartitionerOK with default headers values
func NewFindConfigPartitionerOK() *FindConfigPartitionerOK {
	return &FindConfigPartitionerOK{}
}

/*
FindConfigPartitionerOK handles this case with default header values.

Config value
*/
type FindConfigPartitionerOK struct {
	Payload string
}

func (o *FindConfigPartitionerOK) GetPayload() string {
	return o.Payload
}

func (o *FindConfigPartitionerOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewFindConfigPartitionerDefault creates a FindConfigPartitionerDefault with default headers values
func NewFindConfigPartitionerDefault(code int) *FindConfigPartitionerDefault {
	return &FindConfigPartitionerDefault{
		_statusCode: code,
	}
}

/*
FindConfigPartitionerDefault handles this case with default header values.

unexpected error
*/
type FindConfigPartitionerDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the find config partitioner default response
func (o *FindConfigPartitionerDefault) Code() int {
	return o._statusCode
}

func (o *FindConfigPartitionerDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *FindConfigPartitionerDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *FindConfigPartitionerDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}
