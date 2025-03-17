package services

import (
	"context"
	"fmt"
	"log"
	"time"

	userspb "2k4sm/grpc-crud/proto/users"
	"2k4sm/grpc-crud/src/models"
	"2k4sm/grpc-crud/src/repositories"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	userRepo repositories.UserRepository
	userspb.UnimplementedUsersServer
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (us *UserService) CreateUser(ctx context.Context, req *userspb.UserRequest) (*userspb.UserResponse, error) {
	if req.Email == "" || req.FirstName == "" || req.LastName == "" || req.PhNumber == "" || req.Dob == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Input: email, first_name, last_name and ph_number are required")
	}

	dateString := req.GetDob()
	parsedDate, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to parse date string: %v", err))
	}

	newUser := &models.User{
		Email:     req.GetEmail(),
		FirstName: req.GetFirstName(),
		PhNumber:  req.GetPhNumber(),
		LastName:  req.GetLastName(),
		Gender:    req.GetGender().String(),
		Dob:       parsedDate,
		Access:    req.GetAccess().String(),
	}

	created, err := us.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Error creating user: %v", err))
	}

	if !created {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("User Already Exists with email id: %s", req.GetEmail()))
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

func (us *UserService) GetUser(ctx context.Context, req *userspb.GetUserRequest) (*userspb.UserResponse, error) {
	if req.Email == nil && req.PhNumber == nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid Input: email or ph_number required")
	}

	var user *models.User
	var err error

	if req.Email != nil && req.PhNumber != nil {
		user, err = us.userRepo.GetUserByEmailAndPhone(ctx, *req.Email, *req.PhNumber)
	} else if req.Email != nil {
		user, err = us.userRepo.GetUserByEmail(ctx, *req.Email)
	} else if req.PhNumber != nil {
		user, err = us.userRepo.GetUserByPhone(ctx, *req.PhNumber)
	}

	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("User Not Found: %v", err))
	}

	if user.Access == "BLOCKED" {
		return nil, status.Error(codes.PermissionDenied, "User Access Blocked")
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

func (us *UserService) BlockUser(ctx context.Context, req *userspb.UserAccessUpdateRequest) (*userspb.UserResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Input: email required")
	}

	user, err := us.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("User not found: %v", err))
	}

	err = us.userRepo.UpdateUserAccess(ctx, req.Email, "BLOCKED")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error blocking user: %v", err))
	}

	user.Access = "BLOCKED"

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

func (us *UserService) UnblockUser(ctx context.Context, req *userspb.UserAccessUpdateRequest) (*userspb.UserResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Input: email required")
	}

	user, err := us.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("User not found: %v", err))
	}

	err = us.userRepo.UpdateUserAccess(ctx, req.Email, "UNBLOCKED")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error unblocking user: %v", err))
	}

	user.Access = "UNBLOCKED"

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

func (us *UserService) UpdateUser(ctx context.Context, req *userspb.UserRequest) (*userspb.UserResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Input: email is required")
	}

	existingUser, err := us.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("User not found: %v", err))
	}

	if existingUser.Access == "BLOCKED" {
		return nil, status.Error(codes.PermissionDenied, "User Access Blocked")
	}

	updatedUser := &models.User{
		Email: req.GetEmail(),
	}

	fieldsToUpdate := []string{}

	if req.FirstName != "" {
		updatedUser.FirstName = req.GetFirstName()
		fieldsToUpdate = append(fieldsToUpdate, "first_name")
	}

	if req.LastName != "" {
		updatedUser.LastName = req.GetLastName()
		fieldsToUpdate = append(fieldsToUpdate, "last_name")
	}

	if req.PhNumber != "" {
		updatedUser.PhNumber = req.GetPhNumber()
		fieldsToUpdate = append(fieldsToUpdate, "ph_number")
	}

	if req.Gender.String() != "" {
		updatedUser.Gender = req.GetGender().String()
		fieldsToUpdate = append(fieldsToUpdate, "gender")
	}

	if req.Dob != "" {
		parsedDate, err := time.Parse("2006-01-02", req.Dob)
		if err == nil {
			updatedUser.Dob = parsedDate
			fieldsToUpdate = append(fieldsToUpdate, "dob")
		}
	}

	if req.Access.String() != "" {
		updatedUser.Access = req.GetAccess().String()
		fieldsToUpdate = append(fieldsToUpdate, "access")
	}

	if len(fieldsToUpdate) == 0 {
		return nil, status.Error(codes.InvalidArgument, "No fields to update")
	}

	err = us.userRepo.UpdateUser(ctx, updatedUser, fieldsToUpdate)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error updating user: %v", err))
	}

	updatedUserData, err := us.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Error retrieving updated user: %v", err))
	}

	log.Println("User updated successfully")
	return &userspb.UserResponse{
		Email:     updatedUserData.Email,
		FirstName: updatedUserData.FirstName,
		LastName:  updatedUserData.LastName,
		PhNumber:  updatedUserData.PhNumber,
		Gender:    models.GenderStrToGender(updatedUserData.Gender),
		Dob:       updatedUserData.Dob.Format(time.RFC3339),
		Access:    models.AccessStrToAccess(updatedUserData.Access),
	}, nil
}

func (us *UserService) UpdatePhoneOrEmail(ctx context.Context, req *userspb.UpdatePhoneOrEmailRequest) (*userspb.UserResponse, error) {
	if req.GetCurrEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "Current email is required")
	}

	if req.GetNewEmail() == "" && req.GetNewPhNumber() == "" {
		return nil, status.Error(codes.InvalidArgument, "Either new email or new phone number must be provided")
	}

	user, err := us.userRepo.GetUserByEmail(ctx, req.GetCurrEmail())
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("User not found: %v", err))
	}

	if user.Access == "BLOCKED" {
		return nil, status.Error(codes.PermissionDenied, "User Access Blocked")
	}

	if req.GetNewPhNumber() != "" && req.GetNewEmail() == "" {
		updatedUser := *user
		updatedUser.PhNumber = req.GetNewPhNumber()

		err = us.userRepo.UpdateUser(ctx, &updatedUser, []string{"ph_number"})
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Error updating phone number: %v", err))
		}

		user.PhNumber = req.GetNewPhNumber()
	} else if req.GetNewEmail() != "" {
		_, err = us.userRepo.GetUserByEmail(ctx, req.GetNewEmail())
		if err == nil {
			return nil, status.Error(codes.AlreadyExists, "User with email already exists")
		}

		newUser := *user
		newUser.Email = req.GetNewEmail()

		if req.GetNewPhNumber() != "" {
			newUser.PhNumber = req.GetNewPhNumber()
		}

		created, err := us.userRepo.CreateUser(ctx, &newUser)
		if err != nil || !created {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Error creating new user record: %v", err))
		}

		err = us.userRepo.DeleteUser(ctx, req.GetCurrEmail())
		if err != nil {
			log.Printf("Warning: Failed to delete old user record: %v", err)
		}

		*user = newUser
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
