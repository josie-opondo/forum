package authrepo

import (
	//"database/sql"
	"testing"
	"time"
)

func TestCreateUser(t *testing.T) {
	db := CreateDb()
	userrepo := &UserRepository{DB: db.Db}
	user := &User{Email: "h@R.COM", Password: "Naaahshshs786$", UserID: "2", UserName: "Abas", CreatedAt: time.Now(), UpdatedAt: time.Now().Add(1 * time.Hour)}
	err := userrepo.CreateUser(user)
	if err == nil {
		t.Errorf("expected %v got %v", nil, err)
	}
}

func TestGetUserByEmail(t *testing.T) {
	db := CreateDb()
	userrepo := &UserRepository{DB: db.Db}
	_, err := userrepo.GetUserByEmail("yut@fmail.com")
	if err != nil {
		t.Errorf("expected %v got %v", nil, err)
	}
}

func TestDeleteUser(t *testing.T) {
	db := CreateDb()
	userrepo := &UserRepository{DB: db.Db}
	err := userrepo.DeleteUser("")
	if err != nil {
		t.Errorf("expected %v got %v", nil, err)
	}
}

func TestGetUserByID(t *testing.T) {
}

func TestUpdateUser(t *testing.T) {
}
