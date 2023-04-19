package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/LorezV/go-diploma.git/internal/database"
	"github.com/LorezV/go-diploma.git/internal/models"
	"github.com/jackc/pgx/v4"
)

type UserRepository struct {
	db *database.Database
}

func NewUserRepository(db *database.Database) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (ur *UserRepository) Create(ctx context.Context, login, hashedPassword string) error {
	sql := `INSERT INTO "public"."user" (login, password) VALUES ($1, $2);`

	if _, err := ur.db.Exec(ctx, sql, login, hashedPassword); err != nil {
		return err
	}

	return nil
}

func (ur *UserRepository) Update(ctx context.Context, user *models.User) error {
	sql := `UPDATE "public"."user" SET balance = $1 WHERE id = $2;`

	if _, err := ur.db.Exec(ctx, sql, user.Balance, user.ID); err != nil {
		return err
	}

	return nil
}

func (ur *UserRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
	return ur.flexibleFind(ctx, "id", id)
}

func (ur *UserRepository) FindByLogin(ctx context.Context, login string) (*models.User, error) {
	return ur.flexibleFind(ctx, "login", login)
}

func (ur *UserRepository) flexibleFind(ctx context.Context, column string, value interface{}) (*models.User, error) {
	sql := `SELECT id, login, password, balance, created_at, updated_at FROM "public"."user" WHERE %s = $1;`

	user := new(models.User)

	row := ur.db.QueryRow(ctx, fmt.Sprintf(sql, column), value)
	if err := ur.scanUser(row, user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (ur *UserRepository) scanUser(row pgx.Row, user *models.User) error {
	if err := row.Scan(&user.ID, &user.Login, &user.Password, &user.Balance, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return err
	}

	return nil
}
