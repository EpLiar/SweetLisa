package controller

import (
	"context"
	"fmt"
	"github.com/e14914c0-6759-480d-be89-66b7b7676451/SweetLisa/common"
	"github.com/e14914c0-6759-480d-be89-66b7b7676451/SweetLisa/model"
	"github.com/e14914c0-6759-480d-be89-66b7b7676451/SweetLisa/pkg/log"
	"github.com/e14914c0-6759-480d-be89-66b7b7676451/SweetLisa/service"
	"github.com/gin-gonic/gin"
	"time"
)

// PostRegister registers a server
func PostRegister(c *gin.Context) {
	var req model.Server
	if err := c.ShouldBindJSON(&req); err != nil || c.Param("Ticket") != req.Ticket {
		common.ResponseBadRequestError(c)
		return
	}
	// required info
	if req.Host == "" ||
		req.Port == 0 ||
		!req.Argument.Protocol.Valid() ||
		req.Name == "" {
		common.ResponseBadRequestError(c)
		return
	}
	// verify the server ticket
	ticObj, err := service.GetValidTicketObj(nil, req.Ticket)
	if err != nil {
		common.ResponseError(c, err)
		return
	}
	switch ticObj.Type {
	case model.TicketTypeServer, model.TicketTypeRelay:
	default:
		common.ResponseBadRequestError(c)
		return
	}
	go func(req model.Server, chatIdentifier string) {
		// waiting for the starting of BitterJohn
		time.Sleep(5 * time.Second)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		// ping test
		if err := service.Ping(ctx, req); err != nil {
			err = fmt.Errorf("unreachable: %w", err)
			log.Warn("failed to register: %v", err)
			return
		}
		// register
		if err := service.RegisterServer(nil, req); err != nil {
			return
		}
		if err := service.SyncPassagesByChatIdentifier(nil, ctx, chatIdentifier); err != nil {
			return
		}
	}(req, ticObj.ChatIdentifier)
	log.Info("Received a register request from %v: Chat: %v, Name: %v, Type: %v", c.ClientIP(), req.Name, ticObj.ChatIdentifier, ticObj.Type)
	passages := service.GetPassagesByServer(nil, req.Ticket)
	log.Trace("register: %v", passages)
	common.ResponseSuccess(c, passages)
}
