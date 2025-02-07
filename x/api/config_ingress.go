package api

import (
	"fmt"
	"github.com/168yy/netx/x/app"
	"net/http"
	"strings"

	"github.com/168yy/netx/x/config"
	parser "github.com/168yy/netx/x/config/parsing/ingress"
	"github.com/gin-gonic/gin"
)

// swagger:parameters createIngressRequest
type createIngressRequest struct {
	// in: body
	Data config.IngressConfig `json:"data"`
}

// successful operation.
// swagger:response createIngressResponse
type createIngressResponse struct {
	Data Response
}

func createIngress(ctx *gin.Context) {
	// swagger:route POST /config/ingresses Ingress createIngressRequest
	//
	// Create a new ingress, the name of the ingress must be unique in ingress list.
	//
	//     Security:
	//       basicAuth: []
	//
	//     Responses:
	//       200: createIngressResponse

	var req createIngressRequest
	ctx.ShouldBindJSON(&req.Data)

	name := strings.TrimSpace(req.Data.Name)
	if name == "" {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeInvalid, "ingress name is required"))
		return
	}
	req.Data.Name = name

	if app.Runtime.IngressRegistry().IsRegistered(name) {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeDup, fmt.Sprintf("ingress %s already exists", name)))
		return
	}

	v := parser.ParseIngress(&req.Data)

	if err := app.Runtime.IngressRegistry().Register(name, v); err != nil {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeDup, fmt.Sprintf("ingress %s already exists", name)))
		return
	}

	config.OnUpdate(func(c *config.Config) error {
		c.Ingresses = append(c.Ingresses, &req.Data)
		return nil
	})

	ctx.JSON(http.StatusOK, Response{
		Msg: "OK",
	})
}

// swagger:parameters updateIngressRequest
type updateIngressRequest struct {
	// in: path
	// required: true
	Ingress string `uri:"ingress" json:"ingress"`
	// in: body
	Data config.IngressConfig `json:"data"`
}

// successful operation.
// swagger:response updateIngressResponse
type updateIngressResponse struct {
	Data Response
}

func updateIngress(ctx *gin.Context) {
	// swagger:route PUT /config/ingresses/{ingress} Ingress updateIngressRequest
	//
	// Update ingress by name, the ingress must already exist.
	//
	//     Security:
	//       basicAuth: []
	//
	//     Responses:
	//       200: updateIngressResponse

	var req updateIngressRequest
	ctx.ShouldBindUri(&req)
	ctx.ShouldBindJSON(&req.Data)

	name := strings.TrimSpace(req.Ingress)

	if !app.Runtime.IngressRegistry().IsRegistered(name) {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeNotFound, fmt.Sprintf("ingress %s not found", name)))
		return
	}

	req.Data.Name = name

	v := parser.ParseIngress(&req.Data)

	app.Runtime.IngressRegistry().Unregister(name)

	if err := app.Runtime.IngressRegistry().Register(name, v); err != nil {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeDup, fmt.Sprintf("ingress %s already exists", name)))
		return
	}

	config.OnUpdate(func(c *config.Config) error {
		for i := range c.Ingresses {
			if c.Ingresses[i].Name == name {
				c.Ingresses[i] = &req.Data
				break
			}
		}
		return nil
	})

	ctx.JSON(http.StatusOK, Response{
		Msg: "OK",
	})
}

// swagger:parameters deleteIngressRequest
type deleteIngressRequest struct {
	// in: path
	// required: true
	Ingress string `uri:"ingress" json:"ingress"`
}

// successful operation.
// swagger:response deleteIngressResponse
type deleteIngressResponse struct {
	Data Response
}

func deleteIngress(ctx *gin.Context) {
	// swagger:route DELETE /config/ingresses/{ingress} Ingress deleteIngressRequest
	//
	// Delete ingress by name.
	//
	//     Security:
	//       basicAuth: []
	//
	//     Responses:
	//       200: deleteIngressResponse

	var req deleteIngressRequest
	ctx.ShouldBindUri(&req)

	name := strings.TrimSpace(req.Ingress)

	if !app.Runtime.IngressRegistry().IsRegistered(name) {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeNotFound, fmt.Sprintf("ingress %s not found", name)))
		return
	}
	app.Runtime.IngressRegistry().Unregister(name)

	config.OnUpdate(func(c *config.Config) error {
		ingresses := c.Ingresses
		c.Ingresses = nil
		for _, s := range ingresses {
			if s.Name == name {
				continue
			}
			c.Ingresses = append(c.Ingresses, s)
		}
		return nil
	})

	ctx.JSON(http.StatusOK, Response{
		Msg: "OK",
	})
}
