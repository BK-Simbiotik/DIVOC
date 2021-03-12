// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
)

// NewGetFacilityProgramScheduleParams creates a new GetFacilityProgramScheduleParams object
//
// There are no default values defined in the spec.
func NewGetFacilityProgramScheduleParams() GetFacilityProgramScheduleParams {

	return GetFacilityProgramScheduleParams{}
}

// GetFacilityProgramScheduleParams contains all the bound params for the get facility program schedule operation
// typically these are obtained from a http.Request
//
// swagger:parameters getFacilityProgramSchedule
type GetFacilityProgramScheduleParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*Id of facility
	  Required: true
	  In: path
	*/
	FacilityID string
	/*Id of program
	  Required: true
	  In: path
	*/
	ProgramID string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetFacilityProgramScheduleParams() beforehand.
func (o *GetFacilityProgramScheduleParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rFacilityID, rhkFacilityID, _ := route.Params.GetOK("facilityId")
	if err := o.bindFacilityID(rFacilityID, rhkFacilityID, route.Formats); err != nil {
		res = append(res, err)
	}

	rProgramID, rhkProgramID, _ := route.Params.GetOK("programId")
	if err := o.bindProgramID(rProgramID, rhkProgramID, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindFacilityID binds and validates parameter FacilityID from path.
func (o *GetFacilityProgramScheduleParams) bindFacilityID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.FacilityID = raw

	return nil
}

// bindProgramID binds and validates parameter ProgramID from path.
func (o *GetFacilityProgramScheduleParams) bindProgramID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.ProgramID = raw

	return nil
}
