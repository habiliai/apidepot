package domain

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

type User struct {
	Model
	soft_delete.DeletedAt `json:"-"`

	AuthUserId string `gorm:"index:users_gotrue_idx,unique,where:deleted_at=0" json:"-"`

	Role UserRole `gorm:"default:'user'"`

	Name                 string
	Description          string
	GithubEmail          string
	GithubUsername       string
	GithubInstallationId int64
	// TODO: 현재는 access token 이 만료되지 않도록 github app 이 구성되어있음
	// 추후 만료되도록 설정시 access token 을 refresh 하도록 구현 필요
	GithubAccessToken string
	MediumLink        string
	AvatarUrl         string
	// TODO: 사용자가 많아질 경우, github api rate limit 초과 방지를 위해 installation access token 을 store 하기(GithubInstallationAccessToken 은 1시간 동안 유효)
	// GithubInstallationAccessToken string
	// GithubInstallationAccessTokenExpiresAt time.Time

	StorageSizeLimit int
}

func (u *User) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(u).Error, "failed to save user %v", u)
}

func (u *User) Delete(db *gorm.DB) error {
	return errors.Wrapf(db.Delete(u).Error, "failed to delete user %v", u)
}

func (u *User) IsSuperuser() bool {
	return u.Role == UserRoleAdmin
}

func (u *User) GetGithubAccessToken(ctx context.Context) (string, error) {
	accessToken := u.GithubAccessToken
	if accessToken != "" {
		return accessToken, nil
	}

	accessToken = helpers.GetGithubToken(ctx)
	if accessToken != "" {
		return accessToken, nil
	}

	return "", errors.Wrapf(tclerrors.ErrUnauthorized, "failed to get github access token")
}
