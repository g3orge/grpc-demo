package main

import (
	"context"
	"errors"
	"log"
	"net"

	"github.com/g3orge/grpc-demo/cache"
	"github.com/g3orge/grpc-demo/inv"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type invServ struct {
	inv.UnimplementedInvServer
}

func isAdmin(ctx context.Context) bool {
	user, ok := ctx.Value("user").(inv.User)
	return ok && user.Admin
}

func (s invServ) CreateUser(ctx context.Context, req *inv.CreateUserRequest) (*inv.CreateUserResponse, error) {
	if !isAdmin(ctx) {
		return nil, errors.New("not admin")
	}
	req.User.Id = uuid.New().String()

	us := inv.User{
		Id:       req.User.Id,
		Email:    req.User.Email,
		Username: req.User.Username,
		Password: req.User.Password,
		Admin:    req.User.Admin,
	}
	// log.Println(us)

	c.Set(req.User.Email, us)

	return &inv.CreateUserResponse{
		Done: "done",
	}, nil
}

func (s invServ) GetAllUsers(ctx context.Context, req *inv.GetUsersRequest) (*inv.GetUsersResponse, error) {
	page := req.Page
	pageSize := req.PageSize
	users := c.GetAll()

	//log.Println(users)
	var us []inv.User

	for i := range users {
		us = append(us, users[i])
	}

	var pUs []*inv.User
	for k, _ := range us {
		pUs = append(pUs, &us[k])
	}

	if page == 0 && pageSize == 0 {
		return &inv.GetUsersResponse{
			Users: pUs,
		}, nil
	}
	startIndex := (page - 1) * pageSize
	endIndex := page * pageSize

	indx := len(pUs)
	if startIndex < 0 || startIndex >= int64(indx) {
		return &inv.GetUsersResponse{}, errors.New("limit page")
	}

	if endIndex > int64(indx) {
		endIndex = int64(indx)
	}

	pagedUsers := pUs[startIndex:endIndex]

	// log.Println(us)
	return &inv.GetUsersResponse{
		Users: pagedUsers,
		Total: int64(indx),
	}, nil
}

func (s invServ) GetUserById(ctx context.Context, req *inv.GetUserByIdRequest) (*inv.GetUserResponse, error) {
	// log.Println("enter gubi")
	us, ok := c.GetById(req.Id)
	if !ok {
		return nil, errors.New("could not find user")
	}

	// log.Println("here")
	// log.Println(us)

	return &inv.GetUserResponse{Users: us}, nil
}

func (s invServ) GetUserByName(ctx context.Context, req *inv.GetUserByNameRequest) (*inv.GetUserResponse, error) {
	us, ok := c.GetByName(req.Name)
	if !ok {
		return nil, errors.New("could not find user")
	}
	return &inv.GetUserResponse{Users: us}, nil
}

func (s invServ) UpdateUser(ctx context.Context, req *inv.UpdateUserRequest) (*inv.CreateUserResponse, error) {
	us, ok := c.GetById(req.User.Id)
	if !ok {
		return nil, errors.New("could not update/find user")
	}

	if !isAdmin(ctx) {
		return nil, errors.New("Update: not admin")
	}

	us.Username = req.Name
	us.Password = req.Password
	c.Set(us.Email, *us)

	// log.Println(us)
	return &inv.CreateUserResponse{Done: "user updated"}, nil
}

func (s invServ) DeleteUser(ctx context.Context, req *inv.DeleteUserRequest) (*inv.CreateUserResponse, error) {
	us, ok := c.GetByName(req.Name)
	if !ok {
		return nil, errors.New("could not find user")
	}

	if !isAdmin(ctx) {
		return nil, errors.New("Delete: not admin")
	}

	c.Delete(us.Email)

	return &inv.CreateUserResponse{Done: "done"}, nil
}

func adminAuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	user, ok := ctx.Value("user").(inv.User)
	if !ok {
		return nil, errors.New("unauthenticated: user not found in context")
	}

	if !user.Admin {
		return nil, errors.New("unauthorized: admin access only")
	}

	return handler(ctx, req)
}

var c = cache.New()

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("cannot create listener: %s", err)
	}
	srvReg := grpc.NewServer(
		grpc.UnaryInterceptor(adminAuthInterceptor),
	)
	service := &invServ{}

	inv.RegisterInvServer(srvReg, service)

	err = srvReg.Serve(lis)
	if err != nil {
		log.Fatalf("impossible to serve: %s", err)
	}
}
