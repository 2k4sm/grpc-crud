package repositories

import (
	"2k4sm/grpc-crud/src/models"
	"context"

	"github.com/scylladb/gocqlx/qb"
	"github.com/scylladb/gocqlx/table"
	"github.com/scylladb/gocqlx/v2"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*models.User, error)
	GetUserByEmailAndPhone(ctx context.Context, email, phone string) (*models.User, error)
	UpdateUserAccess(ctx context.Context, email string, access string) error
	UpdateUser(ctx context.Context, user *models.User, fields []string) error
	DeleteUser(ctx context.Context, email string) error
}

type UserRepositoryImpl struct {
	session *gocqlx.Session
	table   *table.Table
}

func NewUserRepository(session *gocqlx.Session) UserRepository {
	return &UserRepositoryImpl{
		session: session,
		table:   table.New(models.UserMetadata),
	}
}

func (r *UserRepositoryImpl) CreateUser(ctx context.Context, user *models.User) (bool, error) {
	stmt, names := qb.Insert(r.table.Name()).
		Columns("email", "first_name", "last_name", "ph_number", "gender", "dob", "access").
		Unique().
		ToCql()

	executor := r.session.Query(stmt, names).BindStruct(user)
	return executor.ExecCASRelease()
}

func (r *UserRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	stmt, names := qb.Select(r.table.Name()).
		Columns("email", "first_name", "last_name", "ph_number", "gender", "dob", "access").
		Where(qb.Eq("email")).
		ToCql()

	executor := r.session.Query(stmt, names).BindMap(qb.M{"email": email})

	var user models.User
	if err := executor.GetRelease(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepositoryImpl) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	stmt, names := qb.Select(r.table.Name()).
		Columns("email", "first_name", "last_name", "ph_number", "gender", "dob", "access").
		Where(qb.Eq("ph_number")).
		ToCql()

	executor := r.session.Query(stmt, names).BindMap(qb.M{"ph_number": phone})

	var user models.User
	if err := executor.GetRelease(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepositoryImpl) GetUserByEmailAndPhone(ctx context.Context, email, phone string) (*models.User, error) {
	stmt, names := qb.Select(r.table.Name()).
		Columns("email", "first_name", "last_name", "ph_number", "gender", "dob", "access").
		Where(qb.Eq("email"), qb.Eq("ph_number")).
		ToCql()

	executor := r.session.Query(stmt, names).BindMap(qb.M{
		"email":     email,
		"ph_number": phone,
	})

	var user models.User
	if err := executor.GetRelease(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepositoryImpl) UpdateUserAccess(ctx context.Context, email string, access string) error {
	stmt, names := qb.Update(r.table.Name()).
		Set("access").
		Where(qb.Eq("email")).
		ToCql()

	executor := r.session.Query(stmt, names).BindMap(qb.M{
		"access": access,
		"email":  email,
	})

	return executor.ExecRelease()
}

func (r *UserRepositoryImpl) UpdateUser(ctx context.Context, user *models.User, fields []string) error {
	updateBuilder := qb.Update(r.table.Name())

	for _, field := range fields {
		updateBuilder = updateBuilder.Set(field)
	}

	stmt, names := updateBuilder.
		Where(qb.Eq("email")).
		ToCql()

	executor := r.session.Query(stmt, names).BindStruct(user)
	return executor.ExecRelease()
}

func (r *UserRepositoryImpl) DeleteUser(ctx context.Context, email string) error {
	stmt, names := qb.Delete(r.table.Name()).
		Where(qb.Eq("email")).
		ToCql()

	executor := r.session.Query(stmt, names).BindMap(qb.M{"email": email})
	return executor.ExecRelease()
}
