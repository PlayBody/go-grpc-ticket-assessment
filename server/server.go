package server

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	proto "github.com/playbody/train-ticket-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"strings"
)

type TrainServer struct {
	proto.UnimplementedTrainServiceServer
	Conf        *TrainConfig
	logger      logr.Logger
	receipts    []map[string][]*proto.User // path, section, seat, userinfo
	flagRoute   map[string]int32
	flagSection map[string]string
	flagSeat    map[string]int32
}

func (s *TrainServer) InitServer() {
	s.logger = logr.Logger{}
	s.flagSection = map[string]string{}
	s.flagSeat = map[string]int32{}
	s.flagRoute = map[string]int32{}
	s.receipts = make([]map[string][]*proto.User, len(s.Conf.Routes))
	for routeIndex := range s.Conf.Routes {
		s.receipts[routeIndex] = make(map[string][]*proto.User)
		for _, sec := range s.Conf.Sections {
			s.receipts[routeIndex][sec] = make([]*proto.User, s.Conf.SeatCount)
		}
	}
}

func (s *TrainServer) AuthUser(_ context.Context, req *proto.AuthRequest) (*proto.AuthResponse, error) {
	var capabilities []string
	for _, user := range SConfig.RoleUsers {
		if user.Email == req.Email {
			capabilities = user.Capabilities
			break
		}
	}
	token, err := GenerateJWT(req.Email, capabilities)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Cannot generate jwt token %v", err)
	}
	return &proto.AuthResponse{Token: token}, nil
}

func (s *TrainServer) GetAllRoutes(_ context.Context, _ *proto.RouteRequest) (*proto.RouteResponse, error) {
	resp := &proto.RouteResponse{
		Routes: make([]*proto.Route, 0),
	}
	for _, route := range s.Conf.Routes {
		resp.Routes = append(resp.Routes, &proto.Route{
			From:  route.From,
			To:    route.To,
			Price: route.Price,
		})
	}
	s.logger.Info("GetAllRoutes", resp)
	return resp, nil
}

func (s *TrainServer) PurchaseTicket(_ context.Context, req *proto.PurchaseRequest) (*proto.PurchaseResponse, error) {
	if _, err := isValidUser(req.User); err != nil {
		return nil, err
	}
	if ok := s.isAlreadyPurchased(req.User.Email); ok {
		return nil, fmt.Errorf("already purchased")
	}
	if index, err := s.getRouteIndex(req.From, req.To, req.Price); err == nil {
		if sec, seat, err := s.findEmptySeat(index); err == nil {
			s.receipts[index][sec][seat] = req.User
			s.flagSeat[req.User.Email] = seat
			s.flagSection[req.User.Email] = sec
			s.flagRoute[req.User.Email] = int32(index)

			resp := &proto.PurchaseResponse{
				Section: sec,
				Seat:    seat,
				Route:   int32(index),
				Message: "Ticket purchased successfully",
			}
			s.logger.Info("PurchaseTicket", resp)
			return resp, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (s *TrainServer) GetReceipt(ctx context.Context, req *proto.ReceiptRequest) (*proto.ReceiptResponse, error) {
	if err := AuthCheck(ctx, req.Email, CapAdmin, CapRead); err != nil {
		return nil, err
	}
	if resp := s.getReceipt(req.Email); resp == nil {
		return nil, fmt.Errorf("no receipt found for email: %v", req.Email)
	} else {
		s.logger.Info("GetReceipt", resp)
		return resp, nil
	}
}

func (s *TrainServer) GetUsersBySection(ctx context.Context, req *proto.SectionRequest) (*proto.SectionResponse, error) {
	if err := AuthCheck(ctx, "", CapAdmin, CapRead); err != nil {
		return nil, err
	}
	seats := make([]*proto.Seat, 0)
	users, ok := s.receipts[req.Route][req.Section]
	if !ok {
		return nil, fmt.Errorf("invalid section: %s", req.Section)
	}
	for i, user := range users {
		if user != nil {
			seats = append(seats, &proto.Seat{
				User: user,
				Seat: int32(i),
			})
		}
	}
	resp := &proto.SectionResponse{Seats: seats}
	s.logger.Info("GetUsersBySection", resp)
	return resp, nil
}

func (s *TrainServer) RemoveUser(ctx context.Context, req *proto.RemoveUserRequest) (*proto.RemoveUserResponse, error) {
	if err := AuthCheck(ctx, req.Email, CapAdmin, CapWrite); err != nil {
		return nil, err
	}
	section, ok := s.flagSection[req.Email]
	if !ok {
		return nil, fmt.Errorf("no user found for email: %v", req.Email)
	}
	seat := s.flagSeat[req.Email]
	route := s.flagRoute[req.Email]
	s.receipts[route][section][seat] = nil
	delete(s.flagSeat, req.Email)
	delete(s.flagSection, req.Email)
	delete(s.flagRoute, req.Email)
	resp := &proto.RemoveUserResponse{
		Route:   route,
		Seat:    seat,
		Section: section,
		Message: "User removed successfully",
	}
	s.logger.Info("RemoveUser", resp)
	return resp, nil
}

func (s *TrainServer) ModifySeat(ctx context.Context, req *proto.ModifySeatRequest) (*proto.ModifySeatResponse, error) {
	if err := AuthCheck(ctx, req.Email, CapAdmin, CapWrite); err != nil {
		return nil, err
	}
	section, ok := s.flagSection[req.Email]
	if !ok {
		return nil, fmt.Errorf("no user found for email: %v", req.Email)
	}
	route := s.flagRoute[req.Email]
	oldSeat := s.flagSeat[req.Email]
	if s.receipts[route][section][req.Seat] != nil {
		return nil, fmt.Errorf("new seat is already occupied")
	}
	user := s.receipts[route][section][oldSeat]

	s.receipts[route][section][oldSeat] = nil
	s.receipts[route][section][req.Seat] = user
	s.flagSeat[req.Email] = req.Seat

	resp := &proto.ModifySeatResponse{
		Message: "Seat modified successfully",
	}
	s.logger.Info("ModifySeat", resp)
	return resp, nil
}

func (s *TrainServer) getRouteIndex(from string, to string, price int32) (int, error) {
	for index, data := range s.Conf.Routes {
		if data.From == from && data.To == to {
			if data.Price <= price {
				return index, nil
			} else {
				return index, fmt.Errorf("you must pay more money")
			}
		}
	}
	return -1, fmt.Errorf("cannot find route")
}

func (s *TrainServer) findEmptySeat(routeIndex int) (string, int32, error) {
	sectionCount := len(s.Conf.Sections)
	x := rand.Intn(sectionCount)
	for i := 0; i < sectionCount; i++ {
		sec := s.Conf.Sections[(i+x)%sectionCount]
		y := rand.Intn(s.Conf.SeatCount)
		for j := 0; j < s.Conf.SeatCount; j++ {
			number := (j + y) % s.Conf.SeatCount
			if s.receipts[routeIndex][sec][number] == nil || s.receipts[routeIndex][sec][number].Email == "" {
				return sec, int32(number), nil
			}
		}
	}
	return "", -1, fmt.Errorf("cannot find empty seat")
}

func (s *TrainServer) isAlreadyPurchased(email string) bool {
	value, is := s.flagSeat[email]
	if is == true && value != -1 {
		return true
	}
	return false
}

func (s *TrainServer) getReceipt(email string) *proto.ReceiptResponse {
	seat, is := s.flagSeat[email]
	if is && seat >= 0 {
		sec, is := s.flagSection[email]
		if is && len(strings.TrimSpace(sec)) > 0 {
			route := s.flagRoute[email]
			return &proto.ReceiptResponse{
				User:    s.receipts[route][sec][seat],
				From:    s.Conf.Routes[route].From,
				To:      s.Conf.Routes[route].To,
				Price:   s.Conf.Routes[route].Price,
				Section: sec,
				Seat:    seat,
			}
		}
		return nil
	}
	return nil
}
