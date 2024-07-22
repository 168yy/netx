package api

import (
	"fmt"
	"github.com/168yy/netx/x/app"
	"net/http"
	"strings"

	"github.com/168yy/netx/x/config"
	parser "github.com/168yy/netx/x/config/parsing/router"
	"github.com/gin-gonic/gin"
)

// swagger:parameters createRouterRequest
type createRouterRequest struct {
	// in: body
	Data config.RouterConfig `json:"data"`
}

// successful operation.
// swagger:response createRouterResponse
type createRouterResponse struct {
	Data Response
}

func createRouter(ctx *gin.Context) {
	// swagger:route POST /config/routers Router createRouterRequest
	//
	// Create a new router, the name of the router must be unique in router list.
	//
	//     Security:
	//       basicAuth: []
	//
	//     Responses:
	//       200: createRouterResponse

	var req createRouterRequest
	ctx.ShouldBindJSON(&req.Data)

	name := strings.TrimSpace(req.Data.Name)
	if name == "" {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeInvalid, "router name is required"))
		return
	}
	req.Data.Name = name

	if app.Runtime.RouterRegistry().IsRegistered(name) {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeDup, fmt.Sprintf("router %s already exists", name)))
		return
	}

	v := parser.ParseRouter(&req.Data)

	if err := app.Runtime.RouterRegistry().Register(name, v); err != nil {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeDup, fmt.Sprintf("router %s already exists", name)))
		return
	}

	config.OnUpdate(func(c *config.Config) error {
		c.Routers = append(c.Routers, &req.Data)
		return nil
	})

	ctx.JSON(http.StatusOK, Response{
		Msg: "OK",
	})
}

// swagger:parameters updateRouterRequest
type updateRouterRequest struct {
	// in: path
	// required: true
	Router string `uri:"router" json:"router"`
	// in: body
	Data config.RouterConfig `json:"data"`
}

// successful operation.
// swagger:response updateRouterResponse
type updateRouterResponse struct {
	Data Response
}

func updateRouter(ctx *gin.Context) {
	// swagger:route PUT /config/routers/{router} Router updateRouterRequest
	//
	// Update router by name, the router must already exist.
	//
	//     Security:
	//       basicAuth: []
	//
	//     Responses:
	//       200: updateRouterResponse

	var req updateRouterRequest
	ctx.ShouldBindUri(&req)
	ctx.ShouldBindJSON(&req.Data)

	name := strings.TrimSpace(req.Router)

	if !app.Runtime.RouterRegistry().IsRegistered(name) {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeNotFound, fmt.Sprintf("router %s not found", name)))
		return
	}

	req.Data.Name = name

	v := parser.ParseRouter(&req.Data)

	app.Runtime.RouterRegistry().Unregister(name)

	if err := app.Runtime.RouterRegistry().Register(name, v); err != nil {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeDup, fmt.Sprintf("router %s already exists", name)))
		return
	}

	config.OnUpdate(func(c *config.Config) error {
		for i := range c.Routers {
			if c.Routers[i].Name == name {
				c.Routers[i] = &req.Data
				break
			}
		}
		return nil
	})

	ctx.JSON(http.StatusOK, Response{
		Msg: "OK",
	})
}

// swagger:parameters deleteRouterRequest
type deleteRouterRequest struct {
	// in: path
	// required: true
	Router string `uri:"router" json:"router"`
}

// successful operation.
// swagger:response deleteRouterResponse
type deleteRouterResponse struct {
	Data Response
}

func deleteRouter(ctx *gin.Context) {
	// swagger:route DELETE /config/routers/{router} Router deleteRouterRequest
	//
	// Delete router by name.
	//
	//     Security:
	//       basicAuth: []
	//
	//     Responses:
	//       200: deleteRouterResponse

	var req deleteRouterRequest
	ctx.ShouldBindUri(&req)

	name := strings.TrimSpace(req.Router)

	if !app.Runtime.RouterRegistry().IsRegistered(name) {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeNotFound, fmt.Sprintf("router %s not found", name)))
		return
	}
	app.Runtime.RouterRegistry().Unregister(name)

	config.OnUpdate(func(c *config.Config) error {
		routeres := c.Routers
		c.Routers = nil
		for _, s := range routeres {
			if s.Name == name {
				continue
			}
			c.Routers = append(c.Routers, s)
		}
		return nil
	})

	ctx.JSON(http.StatusOK, Response{
		Msg: "OK",
	})
}
