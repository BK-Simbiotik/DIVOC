// Code generated by go-swagger; DO NOT EDIT.

package side_effects

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetSideEffectsMetadataHandlerFunc turns a function with the right signature into a get side effects metadata handler
type GetSideEffectsMetadataHandlerFunc func(GetSideEffectsMetadataParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetSideEffectsMetadataHandlerFunc) Handle(params GetSideEffectsMetadataParams) middleware.Responder {
	return fn(params)
}

// GetSideEffectsMetadataHandler interface for that can handle valid get side effects metadata params
type GetSideEffectsMetadataHandler interface {
	Handle(GetSideEffectsMetadataParams) middleware.Responder
}

// NewGetSideEffectsMetadata creates a new http.Handler for the get side effects metadata operation
func NewGetSideEffectsMetadata(ctx *middleware.Context, handler GetSideEffectsMetadataHandler) *GetSideEffectsMetadata {
	return &GetSideEffectsMetadata{Context: ctx, Handler: handler}
}

/* GetSideEffectsMetadata swagger:route GET /sideEffects sideEffects getSideEffectsMetadata

Get Side Effects Metadata

*/
type GetSideEffectsMetadata struct {
	Context *middleware.Context
	Handler GetSideEffectsMetadataHandler
}

func (o *GetSideEffectsMetadata) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetSideEffectsMetadataParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
