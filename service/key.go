package service

import (
	"errors"
	"github.com/boltdb/bolt"
	"github.com/e14914c0-6759-480d-be89-66b7b7676451/SweetLisa/common"
	"github.com/e14914c0-6759-480d-be89-66b7b7676451/SweetLisa/db"
	"github.com/e14914c0-6759-480d-be89-66b7b7676451/SweetLisa/model"
	"github.com/e14914c0-6759-480d-be89-66b7b7676451/SweetLisa/pkg/log"
	jsoniter "github.com/json-iterator/go"
	"strconv"
)

func GetPassagesByServer(tx *bolt.Tx, server model.Server) (passages []model.Passage) {
	// server could be Server or Relay
	f := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(model.BucketTicket))
		if bkt == nil {
			return nil
		}
		// get server chatIdentifier
		bServerTicket := bkt.Get([]byte(server.Ticket))
		var serverTicket model.Ticket
		if err := jsoniter.Unmarshal(bServerTicket, &serverTicket); err != nil {
			return err
		}
		chatIdentifier := serverTicket.ChatIdentifier
		// generate all user/relay passages in this chat
		var userTickets []string
		var servers []model.Server
		var relays []model.Server
		_ = bkt.ForEach(func(k, v []byte) error {
			var ticket model.Ticket
			if err := jsoniter.Unmarshal(v, &ticket); err != nil {
				// do not stop the iter
				return nil
			}
			if ticket.ChatIdentifier != chatIdentifier {
				return nil
			}
			if common.Expired(ticket.ExpireAt) {
				return nil
			}
			switch ticket.Type {
			case model.TicketTypeUser:
				userTickets = append(userTickets, ticket.Ticket)
			case model.TicketTypeServer:
				if serverTicket.Type == model.TicketTypeRelay {
					svr, err := GetServerByTicket(tx, ticket.Ticket)
					if err != nil {
						if !errors.Is(err, db.ErrKeyNotFound) {
							log.Warn("GetPassagesByServer: cannot get server by ticket: %v: %v", ticket.Ticket, err)
						}
						return nil
					}
					servers = append(servers, svr)
				}
			case model.TicketTypeRelay:
				if serverTicket.Type == model.TicketTypeServer {
					relay, err := GetServerByTicket(tx, ticket.Ticket)
					if err != nil {
						if !errors.Is(err, db.ErrKeyNotFound) {
							log.Warn("GetPassagesByServer: cannot get server by ticket: %v: %v", ticket.Ticket, err)
						}
						return nil
					}
					relays = append(relays, relay)
				}
			}
			return nil
		})
		switch serverTicket.Type {
		case model.TicketTypeServer:
			for _, ticket := range userTickets {
				passages = append(passages, model.Passage{
					In: model.In{Argument: server.GetUserArgument(ticket)},
				})
			}
			for _, relay := range relays {
				for _, userTicket := range userTickets {
					passages = append(passages, model.Passage{
						In: model.In{
							From:     relay.Name,
							Argument: relay.GetRelayUserArgument(userTicket, server),
						},
					})
				}
			}
		case model.TicketTypeRelay:
			for _, svr := range servers {
				for _, userTicket := range userTickets {
					arg := server.GetRelayUserArgument(userTicket, svr)
					passages = append(passages, model.Passage{
						In: model.In{Argument: arg}, // TODO: other Protocols and arguments
						Out: &model.Out{
							To:       svr.Name,
							Host:     svr.Host,
							Port:     strconv.Itoa(svr.Port),
							Argument: arg,
						},
					})
				}
			}
		}
		return nil
	}
	if tx != nil {
		f(tx)
		return passages
	}
	db.DB().View(f)
	return passages
}
