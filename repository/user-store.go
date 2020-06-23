package repository

import (
	"database/sql"
	"github.com/ivartj/kartotek/core"
	entity "github.com/ivartj/kartotek/core/entity"
	"github.com/ivartj/kartotek/util/sqlutil"
)

type UserStore struct {
	db core.DB
}

func NewUserStore(db core.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

func (store *UserStore) Get(id entity.UserID) (*entity.User, error) {
	row := store.db.QueryRow("select * from user where user_id = ?;", id)
	var user entity.User
	err := sqlutil.Row{row}.ScanEntity(&user)
	if err == sql.ErrNoRows {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (store *UserStore) GetByUsername(username string) (*entity.User, error) {
	row := store.db.QueryRow("select * from user where username = ?;", username)
	var user entity.User
	err := sqlutil.Row{row}.ScanEntity(&user)
	if err == sql.ErrNoRows {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (store *UserStore) Update(user *entity.User) error {
	return sqlutil.DB{store.db}.InsertOrReplaceEntity("user", user)
}

func (store *UserStore) Delete(id entity.UserID) error {
	_, err := store.db.Exec("delete from user where user_id = ?;", id)
	return err
}
