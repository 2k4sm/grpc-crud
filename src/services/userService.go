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

	if user.Access == "BLOCKED" {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintln("User Access Blocked"))
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

func (us *userService) BlockUser(ctx context.Context, req *userspb.UserAccessUpdateRequest) (*userspb.UserResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Input: email required")
	}

	stmt, names := qb.Select(us.table.Name()).
		Columns("email", "first_name", "last_name", "ph_number", "gender", "dob", "access").
		Where(qb.Eq("email")).
		ToCql()

	var existingUser models.User

	executor := us.session.Query(stmt, names).BindMap(qb.M{"email": req.Email})

	if err := executor.GetRelease(&existingUser); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("User not found: %v", err))
	}

	stmt, names = qb.Update(us.table.Name()).
		Set("access").
		Where(qb.Eq("email")).
		ToCql()

	executor = us.session.Query(stmt, names).BindMap(qb.M{
		"access": "BLOCKED",
		"email":  req.Email,
	})

	existingUser.Access = "BLOCKED"

	if err := executor.ExecRelease(); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error blocking user: %v", err))
	}

	return &userspb.UserResponse{
		Email:     existingUser.Email,
		FirstName: existingUser.FirstName,
		LastName:  existingUser.LastName,
		PhNumber:  existingUser.PhNumber,
		Gender:    models.GenderStrToGender(existingUser.Gender),
		Dob:       existingUser.Dob.Format(time.RFC3339),
		Access:    models.AccessStrToAccess(existingUser.Access),
	}, nil
}

func (us *userService) UnblockUser(ctx context.Context, req *userspb.UserAccessUpdateRequest) (*userspb.UserResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Input: email required")
	}

	stmt, names := qb.Select(us.table.Name()).
		Columns("email", "first_name", "last_name", "ph_number", "gender", "dob", "access").
		Where(qb.Eq("email")).
		ToCql()

	var existingUser models.User

	executor := us.session.Query(stmt, names).BindMap(qb.M{"email": req.Email})

	if err := executor.GetRelease(&existingUser); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("User not found: %v", err))
	}

	stmt, names = qb.Update(us.table.Name()).
		Set("access").
		Where(qb.Eq("email")).
		ToCql()

	executor = us.session.Query(stmt, names).BindMap(qb.M{
		"access": "UNBLOCKED",
		"email":  req.Email,
	})

	existingUser.Access = "UNBLOCKED"

	if err := executor.ExecRelease(); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error unblocking user: %v", err))
	}

	return &userspb.UserResponse{
		Email:     existingUser.Email,
		FirstName: existingUser.FirstName,
		LastName:  existingUser.LastName,
		PhNumber:  existingUser.PhNumber,
		Gender:    models.GenderStrToGender(existingUser.Gender),
		Dob:       existingUser.Dob.Format(time.RFC3339),
		Access:    models.AccessStrToAccess(existingUser.Access),
	}, nil
}

func (us *userService) UpdateUser(ctx context.Context, req *userspb.UserRequest) (*userspb.UserResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Input: email, first_name, last_name, ph_number, dob, access and gender are required")
	}

	stmt, names := qb.Select(us.table.Name()).
		Columns("email").
		Where(qb.Eq("email")).
		ToCql()

	executor := us.session.Query(stmt, names).BindMap(qb.M{"email": req.Email})

	var existingUser models.User

	if err := executor.GetRelease(&existingUser); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("User not found: %v", err))
	}

	if existingUser.Access == "BLOCKED" {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintln("User Access Blocked"))
	}

	updatedUser := models.User{
		Email: req.GetEmail(),
	}

	updateBuilder := qb.Update(us.table.Name())

	if req.FirstName != "" {
		updatedUser.FirstName = req.GetFirstName()
		updateBuilder = updateBuilder.Set("first_name")
	}

	if req.LastName != "" {
		updatedUser.LastName = req.GetLastName()
		updateBuilder = updateBuilder.Set("last_name")
	}

	if req.PhNumber != "" {
		updatedUser.PhNumber = req.GetPhNumber()
		updateBuilder = updateBuilder.Set("ph_number")
	}

	if req.Gender.String() != "" {
		updatedUser.Gender = req.GetGender().String()
		updateBuilder = updateBuilder.Set("gender")
	}

	if req.Dob != "" {
		parsedDate, err := time.Parse("2006-01-02", req.Dob)
		if err == nil {
			updatedUser.Dob = parsedDate
			updateBuilder = updateBuilder.Set("dob")
		}
	}

	if req.Access.String() != "" {
		updatedUser.Access = req.GetAccess().String()
		updateBuilder = updateBuilder.Set("access")
	}

	updateQuery, names := updateBuilder.
		Where(qb.Eq("email")).
		ToCql()

	executor = us.session.Query(updateQuery, names).BindStruct(updatedUser)

	if err := executor.ExecRelease(); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error updating user: %v", err))
	}

	log.Println("User updated successfully")
	return &userspb.UserResponse{
		Email:     updatedUser.Email,
		FirstName: updatedUser.FirstName,
		LastName:  updatedUser.LastName,
		PhNumber:  updatedUser.PhNumber,
		Gender:    models.GenderStrToGender(updatedUser.Gender),
		Dob:       updatedUser.Dob.Format(time.RFC3339),
		Access:    models.AccessStrToAccess(existingUser.Access),
	}, nil
}

func (us *userService) UpdatePhoneOrEmail(ctx context.Context, req *userspb.UpdatePhoneOrEmailRequest) (*userspb.UserResponse, error) {
	if req.GetCurrEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "Current email is required")
	}

	if req.GetNewEmail() == "" && req.GetNewPhNumber() == "" {
		return nil, status.Error(codes.InvalidArgument, "Either new email or new phone number must be provided")
	}

	stmt, names := qb.Select(us.table.Name()).
		Columns("email", "first_name", "last_name", "ph_number", "gender", "dob", "access").
		Where(qb.Eq("email")).
		ToCql()

	executor := us.session.Query(stmt, names).BindMap(qb.M{"email": req.GetCurrEmail()})
	var user models.User
	if err := executor.GetRelease(&user); err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("User not found: %v", err))
	}

	if user.Access == "BLOCKED" {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintln("User Access Blocked"))
	}

	if req.GetNewPhNumber() != "" && req.GetNewEmail() == "" {
		stmt, names := qb.Update(us.table.Name()).
			Set("ph_number").
			Where(qb.Eq("email")).
			ToCql()

		executor := us.session.Query(stmt, names).BindMap(qb.M{
			"ph_number": req.GetNewPhNumber(),
			"email":     req.GetCurrEmail(),
		})

		if err := executor.ExecRelease(); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Error updating phone number: %v", err))
		}

		user.PhNumber = req.GetNewPhNumber()
	} else if req.GetNewEmail() != "" {
		stmt, names := qb.Select(us.table.Name()).
			Columns("email").
			Where(qb.Eq("email")).
			ToCql()

		executor := us.session.Query(stmt, names).BindMap(qb.M{"email": req.GetNewEmail()})
		var existingUser models.User
		if err := executor.GetRelease(&existingUser); err == nil {
			return nil, status.Error(codes.AlreadyExists, "User with email already exists")
		}

		newUser := user
		newUser.Email = req.GetNewEmail()

		if req.GetNewPhNumber() != "" {
			newUser.PhNumber = req.GetNewPhNumber()
		}

		stmt, names = qb.Insert(us.table.Name()).
			Columns("email", "first_name", "last_name", "ph_number", "gender", "dob", "access").
			ToCql()

		executor = us.session.Query(stmt, names).BindStruct(newUser)
		if err := executor.ExecRelease(); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Error creating new user record: %v", err))
		}

		stmt, names = qb.Delete(us.table.Name()).
			Where(qb.Eq("email")).
			ToCql()

		executor = us.session.Query(stmt, names).BindMap(qb.M{"email": req.GetCurrEmail()})
		if err := executor.ExecRelease(); err != nil {
			log.Printf("Warning: Failed to delete old user record: %v", err)
		}

		user = newUser
	}

	log.Println("User phone/email updated successfully")

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
