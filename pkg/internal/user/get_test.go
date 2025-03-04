package user_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	usertest "github.com/habiliai/apidepot/pkg/internal/user/test"
	"time"
)

func (s *ServiceTestSuite) TestGetUser() {
	signinResp := usertest.SignIn(s.T(), s.gotrue)
	defer usertest.SignOut(s.T(), s.gotrue)
	token := signinResp.AccessToken

	ctx := helpers.WithAuthToken(s.Context(), token)
	ctx = helpers.WithTx(ctx, s.db)

	user, err := s.service.GetUser(ctx)
	s.Require().NoError(err)
	s.NotNil(user)

	s.NotZero(user.ID)
}

func (s *ServiceTestSuite) TestGetStorageUsages() {
	signInResp := usertest.SignIn(s.T(), s.gotrue)
	defer usertest.SignOut(s.T(), s.gotrue)

	user, err := s.service.GetUserByAuthUserId(s.Context(), signInResp.User.ID.String())
	s.Require().NoError(err)

	st := domain.Stack{
		Project: domain.Project{
			OwnerID: user.ID,
			Owner:   *user,
		},
	}
	s.Require().NoError(st.Save(s.db))

	hist1 := domain.StackHistory{
		Model: domain.Model{
			CreatedAt: time.Now().Add(-time.Hour * 24),
		},
		StorageSize: 25,
		Stack:       st,
	}
	s.Require().NoError(s.db.Save(&hist1).Error)

	hist2 := domain.StackHistory{
		StorageSize: 50,
		Stack:       st,
	}
	s.Require().NoError(s.db.Save(&hist2).Error)

	hist3 := domain.StackHistory{
		Model: domain.Model{
			CreatedAt: time.Now().Add(time.Hour * 24),
		},
		StorageSize: 75,
		Stack:       st,
	}
	s.Require().NoError(s.db.Save(&hist3).Error)

	ctx := helpers.WithTx(s.Context(), s.db)
	ctx = helpers.WithAuthToken(ctx, signInResp.AccessToken)

	usages, err := s.service.GetStorageUsagesLatest(ctx)
	s.Require().NoError(err)

	s.Equal(float64(50.0), usages.Average)
	s.Equal(float64(50.0), usages.Overage)
	if s.Len(usages.AverageInPeriod, 3) {
		s.Equal(25.0, usages.AverageInPeriod[0].Average)
		s.Equal(50.0, usages.AverageInPeriod[1].Average)
		s.Equal(75.0, usages.AverageInPeriod[2].Average)
	}
	s.T().Logf("usages: %v", usages)
}
