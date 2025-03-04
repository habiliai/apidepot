package user

import (
	"context"
	"github.com/google/uuid"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	gotruetypes "github.com/supabase-community/gotrue-go/types"
	"gorm.io/gorm"
	"math"
	"time"
)

type (
	// StorageUsages all fields are defined in "bytes"
	StorageUsages struct {
		Average         float64
		Overage         float64
		AverageInPeriod []struct {
			Date    time.Time
			Average float64
		}
	}
)

func (s *service) GetUser(
	ctx context.Context,
) (user *domain.User, err error) {
	token := helpers.GetAuthToken(ctx)
	if token == "" {
		return user, errors.Wrapf(tclerrors.ErrForbidden, "Invalid authorization header")
	}
	gotrueClient := s.gotrueClient.WithToken(helpers.GetAuthToken(ctx))

	getUserResp, err := gotrueClient.GetUser()
	if err != nil {
		return user, errors.Wrapf(err, "failed to get user")
	}

	return s.GetUserByAuthUserId(ctx, getUserResp.ID.String())
}

func (s *service) GetUserByAuthUserId(
	ctx context.Context,
	ownerId string,
) (user *domain.User, err error) {
	getUserResp, err := s.gotrueClient.AdminGetUser(gotruetypes.AdminGetUserRequest{
		UserID: uuid.MustParse(ownerId),
	})
	if err != nil {
		return user, errors.Wrapf(err, "failed to get user")
	}

	if (getUserResp.EmailConfirmedAt == nil && getUserResp.PhoneConfirmedAt == nil) || getUserResp.BannedUntil != nil {
		return user, errors.Wrapf(tclerrors.ErrNotFound, "user not found")
	}

	if r := helpers.GetTx(ctx).Find(&user, "auth_user_id = ?", ownerId); r.Error != nil {
		return user, errors.Wrapf(r.Error, "failed to find user by auth user id")
	} else if r.RowsAffected > 0 {
		return user, nil
	}

	if err := helpers.GetTx(ctx).Transaction(func(tx *gorm.DB) error {
		user = &domain.User{
			AuthUserId: ownerId,
		}
		if err := tx.Create(&user).Error; err != nil {
			return errors.Wrapf(err, "failed to create user")
		}

		return nil
	}); err != nil {
		return user, err
	}

	return user, nil
}

func (s *service) GetStorageUsagesLatest(
	ctx context.Context,
) (*StorageUsages, error) {
	user, err := s.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	begin := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := now.Month() + 1
	if begin.Month() == 12 {
		nextMonth = 1
	}
	end := time.Date(now.Year(), nextMonth, 1, 0, 0, 0, 0, time.UTC)

	tx := helpers.GetTx(ctx)

	stackHistoriesStmt := tx.
		Model(&domain.StackHistory{}).
		InnerJoins("Stack").
		InnerJoins("Stack.Project").
		InnerJoins("Stack.Project.Owner", "Stack__Project__Owner.id = ?", user.ID).
		Where("stack_histories.created_at BETWEEN ? AND ?", begin, end)

	usages := &StorageUsages{}
	if err := tx.Table("(?) as a", stackHistoriesStmt).Select("AVG(a.storage_size)").Scan(&usages.Average).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get average storage usage")
	}

	if err := tx.Table(
		"(?) as a",
		stackHistoriesStmt.Select(
			"stack_histories.*, to_char(stack_histories.created_at, 'YYYY-MM-DD') as date",
		),
	).Group("date").
		Select("to_timestamp(date, 'YYYY-MM-DD') as date, AVG(storage_size) as average").
		Order("date ASC").
		Scan(&usages.AverageInPeriod).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get average storage usage in period")
	}

	usages.Overage = math.Abs(float64(user.StorageSizeLimit) - usages.Average)

	return usages, nil
}
