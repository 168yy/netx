package api

import (
	"fmt"
	"github.com/168yy/netx/x/app"
	"net/http"
	"strings"

	"github.com/168yy/netx/x/config"
	parser "github.com/168yy/netx/x/config/parsing/sd"
	"github.com/gin-gonic/gin"
)

// swagger:parameters createSDRequest
type createSDRequest struct {
	// in: body
	Data config.SDConfig `json:"data"`
}

// successful operation.
// swagger:response createSDResponse
type createSDResponse struct {
	Data Response
}

func createSD(ctx *gin.Context) {
	// swagger:route POST /config/sds SD createSDRequest
	//
	// Create a new SD, the name of the SD must be unique in SD list.
	//
	//     Security:
	//       basicAuth: []
	//
	//     Responses:
	//       200: createSDResponse

	var req createSDRequest
	ctx.ShouldBindJSON(&req.Data)

	name := strings.TrimSpace(req.Data.Name)
	if name == "" {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeInvalid, "sd name is required"))
		return
	}
	req.Data.Name = name

	if app.Runtime.SDRegistry().IsRegistered(name) {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeDup, fmt.Sprintf("sd %s already exists", name)))
		return
	}

	v := parser.ParseSD(&req.Data)

	if err := app.Runtime.SDRegistry().Register(name, v); err != nil {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeDup, fmt.Sprintf("sd %s already exists", name)))
		return
	}

	config.OnUpdate(func(c *config.Config) error {
		c.SDs = append(c.SDs, &req.Data)
		return nil
	})

	ctx.JSON(http.StatusOK, Response{
		Msg: "OK",
	})
}

// swagger:parameters updateSDRequest
type updateSDRequest struct {
	// in: path
	// required: true
	SD string `uri:"sd" json:"sd"`
	// in: body
	Data config.SDConfig `json:"data"`
}

// successful operation.
// swagger:response updateSDResponse
type updateSDResponse struct {
	Data Response
}

func updateSD(ctx *gin.Context) {
	// swagger:route PUT /config/sds/{sd} SD updateSDRequest
	//
	// Update SD by name, the SD must already exist.
	//
	//     Security:
	//       basicAuth: []
	//
	//     Responses:
	//       200: updateSDResponse

	var req updateSDRequest
	ctx.ShouldBindUri(&req)
	ctx.ShouldBindJSON(&req.Data)

	name := strings.TrimSpace(req.SD)

	if !app.Runtime.SDRegistry().IsRegistered(name) {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeNotFound, fmt.Sprintf("sd %s not found", name)))
		return
	}

	req.Data.Name = name

	v := parser.ParseSD(&req.Data)

	app.Runtime.SDRegistry().Unregister(name)

	if err := app.Runtime.SDRegistry().Register(name, v); err != nil {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeDup, fmt.Sprintf("sd %s already exists", name)))
		return
	}

	config.OnUpdate(func(c *config.Config) error {
		for i := range c.SDs {
			if c.SDs[i].Name == name {
				c.SDs[i] = &req.Data
				break
			}
		}
		return nil
	})

	ctx.JSON(http.StatusOK, Response{
		Msg: "OK",
	})
}

// swagger:parameters deleteSDRequest
type deleteSDRequest struct {
	// in: path
	// required: true
	SD string `uri:"sd" json:"sd"`
}

// successful operation.
// swagger:response deleteSDResponse
type deleteSDResponse struct {
	Data Response
}

func deleteSD(ctx *gin.Context) {
	// swagger:route DELETE /config/sds/{sd} SD deleteSDRequest
	//
	// Delete SD by name.
	//
	//     Security:
	//       basicAuth: []
	//
	//     Responses:
	//       200: deleteSDResponse

	var req deleteSDRequest
	ctx.ShouldBindUri(&req)

	name := strings.TrimSpace(req.SD)

	if !app.Runtime.SDRegistry().IsRegistered(name) {
		writeError(ctx, NewError(http.StatusBadRequest, ErrCodeNotFound, fmt.Sprintf("sd %s not found", name)))
		return
	}
	app.Runtime.SDRegistry().Unregister(name)

	config.OnUpdate(func(c *config.Config) error {
		sds := c.SDs
		c.SDs = nil
		for _, s := range sds {
			if s.Name == name {
				continue
			}
			c.SDs = append(c.SDs, s)
		}
		return nil
	})

	ctx.JSON(http.StatusOK, Response{
		Msg: "OK",
	})
}
