package services

import (
	"context"
	"fmt"
	"log"
	"time"

	userspb "2k4sm/grpc-crud/proto/users"
	"2k4sm/grpc-crud/src/models"

	"github.com/scylladb/gocqlx/qb"
	"github.com/scylladb/gocqlx/table"
	"github.com/scylladb/gocqlx/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userService struct {
	session *gocqlx.Session
	table   *table.Table
	userspb.UnimplementedUsersServer
}

func NewUserService(session *gocqlx.Session) *userService {
	return &userService{
		session: session,
		table:   table.New(models.UserMetadata),
	}
}

func (us *userService) CreateUser(ctx context.Context, req *userspb.UserRequest) (*userspb.UserResponse, error) {

	if req.Email == "" || req.FirstName == "" || req.LastName == "" || req.PhNumber == "" || req.Dob == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Input: email, first_name, last_name and ph_number are required")
	}

	dateString := req.GetDob()

	parsedDate, err := time.Parse("2006-01-02", dateString)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to parse date string: %v", err))
	}

	newUser := models.User{
		Email:     req.GetEmail(),
		FirstName: req.GetFirstName(),
		PhNumber:  req.GetPhNumber(),
		LastName:  req.GetLastName(),
		Gender:    req.GetGender().String(),
		Dob:       parsedDate,
		Access:    req.GetAccess().String(),
	}

	stmt, names := qb.Insert(us.table.Name()).
		Columns("email", "first_name", "last_name", "ph_number", "gender", "dob", "access").
		Unique().
		ToCql()

	executor := us.session.Query(stmt, names).BindStruct(newUser)

	created, err := executor.ExecCASRelease()

	if err != nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Error creating user: ", err))
	}

	if !created {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("User Already Exists with email id : ", req.GetEmail()))
	}

	log.Println("User Created Successfully")

	return &userspb.UserResponse{
		Email:     newUser.Email,
		PhNumber:  newUser.PhNumber,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Gender:    models.GenderStrToGender(newUser.Gender),
		Dob:       parsedDate.Format(time.RFC3339),
		Access:    models.AccessStrToAccess(newUser.Access),
	}, nil
}

func (us *userService) GetUser(ctx context.Context, req *userspb.GetUserRequest) (*userspb.UserResponse, error) {
	if req.Email == nil && req.PhNumber == nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid Input: email or ph_number required")
	}

	stmt := qb.Select(us.table.Name()).
		Columns("email", "first_name", "last_name", "ph_number", "gender", "dob", "access")

	binding := qb.M{}
	if req.Email != nil && req.PhNumber != nil {
		stmt = stmt.Where(qb.Eq("email"), qb.Eq("ph_number"))
		binding["email"] = req.Email
		binding["ph_number"] = req.PhNumber
	} else if req.Email != nil {
		stmt = stmt.Where(qb.Eq("email"))
		binding["email"] = req.Email
	} else if req.PhNumber != nil {
		stmt = stmt.Where(qb.Eq("ph_number"))
		binding["ph_number"] = req.PhNumber
	}

	cql, names := stmt.ToCql()

	executor := us.session.Query(cql, names).BindMap(binding)

	var user models.User

	if err := executor.GetRelease(&user); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("User Not Found: %v", err))
	}

	log.Println("User Found Successfully")
	return &userspb.UserResponse{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		PhNumber:  user.PhNumber,
		Gender:    models.GenderStrToGender(user.Gender),
		Dob:       user.Dob.Format(time.RFC3339),
		Access:    models.AccessStrToAccess(user.Access),
	}, nil
}
