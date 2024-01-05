package server

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"testing"

	proto "github.com/playbody/train-ticket-service/proto"
	"github.com/stretchr/testify/assert"
)

const (
	section1   = "Section1"
	section2   = "Section2"
	firstName1 = "Draw Ranger"
	lastName1  = "Trax"
	email1     = "user1@example.com"
	firstName2 = "Phantom"
	lastName2  = "Mott"
	email2     = "user2@example.com"
	firstName3 = "Medusa"
	lastName3  = "Gorgon"
	email3     = "user3@example.com"
	firstName4 = "Sniper"
	lastName4  = "Headshot"
	email4     = "user4@example.com"
	firstName5 = "Ok"
	lastName5  = "Cool"
	email5     = "user5@example.com"
	seatCount  = 2
	from1      = "London"
	to1        = "Paris"
	price1     = int32(20)
	from2      = "Osaka"
	to2        = "London"
	price2     = int32(200)
)

var server *TrainServer

func TestMain(m *testing.M) {
	server = &TrainServer{
		Conf: &TrainConfig{
			Routes: []struct {
				From  string `yaml:"from,omitempty"`
				To    string `yaml:"to,omitempty"`
				Price int32  `yaml:"price,omitempty"`
			}{
				{
					From:  from1,
					To:    to1,
					Price: price1,
				},
				{
					From:  from2,
					To:    to2,
					Price: price2,
				},
			},
			Sections:  []string{section1, section2},
			SeatCount: seatCount,
		},
		receipts: []map[string][]*proto.User{
			{
				section1: []*proto.User{
					{
						FirstName: firstName1,
						LastName:  lastName1,
						Email:     email1,
					},
					{
						FirstName: firstName2,
						LastName:  lastName2,
						Email:     email2,
					},
				},
				section2: []*proto.User{
					{
						FirstName: firstName3,
						LastName:  lastName3,
						Email:     email3,
					},
					{},
				},
			},
			{
				section1: []*proto.User{{}, {}},
				section2: []*proto.User{{}, {}},
			},
		},
		flagSection: map[string]string{
			email1: section1,
			email2: section1,
			email3: section2,
		},
		flagRoute: map[string]int32{
			email1: 0,
			email2: 0,
			email3: 0,
		},
		flagSeat: map[string]int32{
			email1: 0,
			email2: 1,
			email3: 0,
		},
	}
	SConfig.Auth.Expire = 3600
	SConfig.Auth.SecretKey = "abc"
	SConfig.RoleUsers = []RoleUser{
		{
			Email: email1,
			Capabilities: []string{
				"admin",
				"read",
				"write",
			},
		},
	}
	m.Run()
}

func TestTrainServer_GetAllRoutes(t *testing.T) {
	t.Run("valid from get all route", func(t *testing.T) {
		req := &proto.RouteRequest{}
		resp, err := server.GetAllRoutes(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Routes, 2)
		assert.Equal(t, from1, resp.Routes[0].From)
		assert.Equal(t, to1, resp.Routes[0].To)
		assert.Equal(t, price1, resp.Routes[0].Price)
		assert.Equal(t, from2, resp.Routes[1].From)
		assert.Equal(t, to2, resp.Routes[1].To)
		assert.Equal(t, price2, resp.Routes[1].Price)
	})
}
func TestTrainServer_PurchaseTicket(t *testing.T) {
	t.Run("valid purchase ticket", func(t *testing.T) {
		req := &proto.PurchaseRequest{
			User: &proto.User{
				FirstName: firstName4,
				LastName:  lastName4,
				Email:     email4,
			},
			From:  from1,
			To:    to1,
			Price: price1,
		}
		resp, err := server.PurchaseTicket(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Ticket purchased successfully", resp.Message)
	})

	t.Run("already purchased ticket", func(t *testing.T) {
		req := &proto.PurchaseRequest{
			User: &proto.User{
				FirstName: firstName1,
				LastName:  lastName1,
				Email:     email1,
			},
			From:  from1,
			To:    to1,
			Price: price1,
		}
		_, err := server.PurchaseTicket(context.Background(), req)
		assert.Error(t, err)
		assert.EqualError(t, err, "already purchased")
	})

	t.Run("invalid user", func(t *testing.T) {
		req := &proto.PurchaseRequest{
			User: &proto.User{
				FirstName: "",
				LastName:  lastName5,
				Email:     email5,
			},
			From:  from1,
			To:    to1,
			Price: price1,
		}
		_, err := server.PurchaseTicket(context.Background(), req)
		assert.Error(t, err)
		assert.EqualError(t, err, "first name must not be empty")
	})

	t.Run("invalid route", func(t *testing.T) {
		req := &proto.PurchaseRequest{
			User: &proto.User{
				FirstName: firstName5,
				LastName:  lastName5,
				Email:     email5,
			},
			From:  "invalid_from",
			To:    "invalid_to",
			Price: price1,
		}
		_, err := server.PurchaseTicket(context.Background(), req)
		assert.Error(t, err)
		assert.EqualError(t, err, "cannot find route")
	})

	t.Run("not enough money", func(t *testing.T) {
		req := &proto.PurchaseRequest{
			User: &proto.User{
				FirstName: firstName5,
				LastName:  lastName5,
				Email:     email5,
			},
			From:  from1,
			To:    to1,
			Price: 10,
		}
		_, err := server.PurchaseTicket(context.Background(), req)
		assert.Error(t, err)
		assert.EqualError(t, err, "you must pay more money")
	})
}

func TestTrainServer_GetReceipt(t *testing.T) {
	t.Run("get receipt success", func(t *testing.T) {
		req1 := &proto.AuthRequest{
			Email: email1,
		}
		resp1, err := server.AuthUser(context.Background(), req1)
		assert.Nil(t, err)
		assert.Greater(t, len(resp1.Token), 0, "Token length must have greater than 0.")

		md := metadata.Pairs("Authorization", resp1.Token)
		ctx := metadata.NewIncomingContext(context.Background(), md)
		var handler grpc.UnaryHandler = func(ctx context.Context, req any) (any, error) {
			resp, err := server.GetReceipt(ctx, req.(*proto.ReceiptRequest))
			return resp, err
		}
		req2 := &proto.ReceiptRequest{
			Email: email1,
		}
		resp, err := ParseJWTMiddleware(ctx, req2, &grpc.UnaryServerInfo{}, handler)

		assert.NotNil(t, resp)
		assert.Nil(t, err)
	})
}
