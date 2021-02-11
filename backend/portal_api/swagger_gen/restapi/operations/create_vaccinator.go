// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/divoc/portal-api/swagger_gen/models"
)

// CreateVaccinatorHandlerFunc turns a function with the right signature into a create vaccinator handler
type CreateVaccinatorHandlerFunc func(CreateVaccinatorParams, *models.JWTClaimBody) middleware.Responder

// Handle executing the request and returning a response
func (fn CreateVaccinatorHandlerFunc) Handle(params CreateVaccinatorParams, principal *models.JWTClaimBody) middleware.Responder {
	return fn(params, principal)
}

// CreateVaccinatorHandler interface for that can handle valid create vaccinator params
type CreateVaccinatorHandler interface {
	Handle(CreateVaccinatorParams, *models.JWTClaimBody) middleware.Responder
}

// NewCreateVaccinator creates a new http.Handler for the create vaccinator operation
func NewCreateVaccinator(ctx *middleware.Context, handler CreateVaccinatorHandler) *CreateVaccinator {
	return &CreateVaccinator{Context: ctx, Handler: handler}
}

/* CreateVaccinator swagger:route POST /vaccinator createVaccinator

Create vaccinator user

*/
type CreateVaccinator struct {
	Context *middleware.Context
	Handler CreateVaccinatorHandler
}

func (o *CreateVaccinator) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewCreateVaccinatorParams()
	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		r = aCtx
	}
	var principal *models.JWTClaimBody
	if uprinc != nil {
		principal = uprinc.(*models.JWTClaimBody) // this is really a models.JWTClaimBody, I promise
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}