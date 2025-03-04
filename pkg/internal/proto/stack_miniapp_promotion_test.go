package proto_test

import (
	"github.com/google/uuid"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/mokiat/gog"
	"google.golang.org/grpc/metadata"
	"time"
)

func (s *ProtoTestSuite) TestGetAllPublicTelegramMiniappPromotions() {
	// Given
	project := domain.Project{
		Owner: domain.User{
			Name: "test-user",
		},
		Name: "test-project",
	}
	s.Require().NoError(s.db.Save(&project).Error)

	stacks := []domain.Stack{
		{
			ProjectID: project.ID,
			Name:      "test-stack-1",
			Hash:      "test-hash-1",
			ServiceTemplate: &domain.ServiceTemplate{
				Name: "test-service-template-1",
			},
		},
		{
			ProjectID: project.ID,
			Name:      "test-stack-2",
			Hash:      "test-hash-2",
			ServiceTemplate: &domain.ServiceTemplate{
				Name: "test-service-template-2",
			},
		},
		{
			ProjectID: project.ID,
			Name:      "test-stack-3",
			Hash:      "test-hash-3",
			ServiceTemplate: &domain.ServiceTemplate{
				Name: "test-service-template-3",
			},
		},
		{
			ProjectID: project.ID,
			Name:      "test-stack-4",
			Hash:      "test-hash-4",
			ServiceTemplate: &domain.ServiceTemplate{
				Name: "test-service-template-4",
			},
		},
		{
			ProjectID: project.ID,
			Name:      "test-stack-5",
			Hash:      "test-hash-5",
			ServiceTemplate: &domain.ServiceTemplate{
				Name: "test-service-template-5",
			},
		},
		{
			ProjectID: project.ID,
			Name:      "test-stack-6",
			Hash:      "test-hash-6",
			ServiceTemplate: &domain.ServiceTemplate{
				Name: "test-service-template-6",
			},
		},
	}
	s.Require().NoError(s.db.Save(&stacks).Error)

	t := time.Now().Add(-2 * time.Hour)
	promotions := []domain.TelegramMiniappPromotion{
		{
			StackID:                stacks[0].ID,
			Link:                   "https://example.com/1",
			AppTitle:               "App 1",
			AppIconImageUrl:        "https://example.com/icon1.png",
			AppDescription:         "Description 1",
			AppScreenshotImageUrls: []string{"https://example.com/screenshot1.png"},
			AppBannerImageUrl:      "https://example.com/banner1.png",
			Views:                  []domain.TelegramMiniappPromotionView{{DeviceID: "test-device-id-1", NumViews: 100}, {DeviceID: "test-device-id-2", NumViews: 200}},
			CreatedAt:              time.Now().Add(-5 * time.Hour),
			Public:                 true,
		},
		{
			StackID:                stacks[1].ID,
			Link:                   "https://example.com/2",
			AppTitle:               "App 2",
			AppIconImageUrl:        "https://example.com/icon2.png",
			AppDescription:         "Description 2",
			AppScreenshotImageUrls: []string{"https://example.com/screenshot2.png"},
			AppBannerImageUrl:      "https://example.com/banner2.png",
			Views:                  []domain.TelegramMiniappPromotionView{{DeviceID: "test-device-id-1", NumViews: 100}},
			CreatedAt:              time.Now().Add(-4 * time.Hour),
			Public:                 true,
		},
		{
			StackID:                stacks[2].ID,
			Link:                   "https://example.com/3",
			AppTitle:               "App 3",
			AppIconImageUrl:        "https://example.com/icon3.png",
			AppDescription:         "Description 3",
			AppScreenshotImageUrls: []string{"https://example.com/screenshot3.png"},
			AppBannerImageUrl:      "https://example.com/banner3.png",
			Views:                  []domain.TelegramMiniappPromotionView{{DeviceID: "test-device-id-1", NumViews: 100}, {DeviceID: "test-device-id-2", NumViews: 200}, {DeviceID: "test-device-id-3", NumViews: 300}},
			CreatedAt:              time.Now().Add(-3 * time.Hour),
			Public:                 true,
		},
		{
			StackID:                stacks[3].ID,
			Link:                   "https://example.com/4",
			AppTitle:               "App 4",
			AppIconImageUrl:        "https://example.com/icon4.png",
			AppDescription:         "Description 4",
			AppScreenshotImageUrls: []string{"https://example.com/screenshot4.png"},
			AppBannerImageUrl:      "https://example.com/banner4.png",
			Views:                  []domain.TelegramMiniappPromotionView{{DeviceID: "test-device-id-1", NumViews: 100}},
			CreatedAt:              t,
			Public:                 true,
		},
		{
			StackID:                stacks[5].ID,
			Link:                   "https://example.com/6",
			AppTitle:               "App 6",
			AppIconImageUrl:        "https://example.com/icon6.png",
			AppDescription:         "Description 6",
			AppScreenshotImageUrls: []string{"https://example.com/screenshot6.png"},
			AppBannerImageUrl:      "https://example.com/banner6.png",
			Views:                  []domain.TelegramMiniappPromotionView{{DeviceID: "test-device-id-1", NumViews: 100}, {DeviceID: "test-device-id-2", NumViews: 200}, {DeviceID: "test-device-id-3", NumViews: 300}, {DeviceID: "test-device-id-4", NumViews: 400}},
			CreatedAt:              t,
			Public:                 true,
		},
		{
			StackID:                stacks[4].ID,
			Link:                   "https://example.com/5",
			AppTitle:               "App 5",
			AppIconImageUrl:        "https://example.com/icon5.png",
			AppDescription:         "Description 5",
			AppScreenshotImageUrls: []string{"https://example.com/screenshot5.png"},
			AppBannerImageUrl:      "https://example.com/banner5.png",
			Views:                  []domain.TelegramMiniappPromotionView{{DeviceID: "test-device-id-1", NumViews: 100}},
			CreatedAt:              time.Now().Add(-1 * time.Hour),
			Public:                 true,
		},
	}
	s.Require().NoError(s.db.Save(&promotions).Error)

	tcc, dispose := s.newClient()
	defer dispose()

	s.Run("When `GetAllPublicTelegramMiniappPromotions` is called order by created_at, Should be returned sorted results", func() {
		// When
		resp1, err := tcc.GetAllPublicTelegramMiniappPromotions(s.Context(), &proto.GetAllPublicTelegramMiniappPromotionsRequest{
			OrderBy: gog.PtrOf(proto.GetAllPublicTelegramMiniappPromotionsRequest_ORDER_BY_CREATED_AT),
			Limit:   gog.PtrOf(int32(2)),
		})
		s.Require().NoError(err)

		// Then
		s.Equal(int32(6), resp1.NumTotal)
		s.Len(resp1.Records, 2)
		s.Equal("App 5", resp1.Records[0].AppTitle)
		s.Equal("App 4", resp1.Records[1].AppTitle)

		resp2, err := tcc.GetAllPublicTelegramMiniappPromotions(s.Context(), &proto.GetAllPublicTelegramMiniappPromotionsRequest{
			OrderBy: gog.PtrOf(proto.GetAllPublicTelegramMiniappPromotionsRequest_ORDER_BY_CREATED_AT),
			Limit:   gog.PtrOf(int32(3)),
			Cursor:  gog.PtrOf(resp1.NextCursor),
		})

		// Then
		s.Equal(int32(6), resp2.NumTotal)
		s.Len(resp2.Records, 3)
		s.Equal("App 6", resp2.Records[0].AppTitle)
		s.Equal("App 3", resp2.Records[1].AppTitle)
		s.Equal("App 2", resp2.Records[2].AppTitle)

		s.Require().NotNil(resp2.Records[2].Stack)
		s.Require().NotNil(resp2.Records[2].Stack.ServiceTemplateId)
	})

	s.Run("When `GetAllPublicTelegramMiniappPromotions` is called order by num_views, Should be returned sorted results", func() {
		// When
		resp1, err := tcc.GetAllPublicTelegramMiniappPromotions(s.Context(), &proto.GetAllPublicTelegramMiniappPromotionsRequest{
			OrderBy: gog.PtrOf(proto.GetAllPublicTelegramMiniappPromotionsRequest_ORDER_BY_NUM_VIEWS),
			Limit:   gog.PtrOf(int32(2)),
		})
		s.Require().NoError(err)

		// Then
		s.Equal(int32(6), resp1.NumTotal)
		s.Len(resp1.Records, 2)
		s.Equal("App 6", resp1.Records[0].AppTitle)
		s.Equal("App 3", resp1.Records[1].AppTitle)

		s.NotEmpty(resp1.NextCursor)
		s.T().Logf("NextCursor: %s", resp1.NextCursor)

		resp2, err := tcc.GetAllPublicTelegramMiniappPromotions(s.Context(), &proto.GetAllPublicTelegramMiniappPromotionsRequest{
			OrderBy: gog.PtrOf(proto.GetAllPublicTelegramMiniappPromotionsRequest_ORDER_BY_NUM_VIEWS),
			Limit:   gog.PtrOf(int32(3)),
			Cursor:  gog.PtrOf(resp1.NextCursor),
		})

		// Then
		s.Equal(int32(6), resp2.NumTotal)
		s.Len(resp2.Records, 3)
		s.Equal("App 1", resp2.Records[0].AppTitle)
		s.Equal("App 2", resp2.Records[1].AppTitle)
		s.Equal("App 4", resp2.Records[2].AppTitle)

		s.Require().NotNil(resp2.Records[2].Stack)
		s.Require().NotNil(resp2.Records[2].Stack.ServiceTemplateId)
	})
}

func (s *ProtoTestSuite) TestGetPublicTelegramMiniappPromotion() {
	// Given
	project := domain.Project{
		Owner: domain.User{
			Name: "test-user",
		},
		Name: "test-project",
	}
	s.Require().NoError(s.db.Save(&project).Error)

	stack := domain.Stack{
		ProjectID: project.ID,
		Name:      "test-stack",
		Hash:      "test-hash",
		ServiceTemplate: &domain.ServiceTemplate{
			Name: "test-service-template",
		},
	}
	s.Require().NoError(s.db.Save(&stack).Error)

	promotion := domain.TelegramMiniappPromotion{
		StackID:                stack.ID,
		Link:                   "https://example.com",
		AppTitle:               "Test App",
		AppIconImageUrl:        "https://example.com/icon.png",
		AppDescription:         "Test Description",
		AppScreenshotImageUrls: []string{"https://example.com/screenshot.png"},
		AppBannerImageUrl:      "https://example.com/banner.png",
		CreatedAt:              time.Now().Add(-1 * time.Hour),
		Public:                 true,
	}
	s.Require().NoError(s.db.Save(&promotion).Error)

	tcc, dispose := s.newClient()
	defer dispose()

	// When
	ctx := metadata.AppendToOutgoingContext(s.Context(), "x-device-id", uuid.NewString())
	resp1, err := tcc.GetPublicTelegramMiniappPromotion(ctx, &proto.GetPublicTelegramMiniappPromotionRequest{
		StackId: int32(promotion.StackID),
	})
	s.Require().NoError(err)
	s.Equal(int32(0), resp1.NumUniqueViews)

	// Then
	s.Equal("Test App", resp1.AppTitle)
	s.Require().NotNil(resp1.Stack)
	s.Equal(stack.ID, uint(resp1.Stack.Id))
	s.Require().NotNil(resp1.Stack.ServiceTemplateId)
	s.Equal(*stack.ServiceTemplateID, uint(*resp1.Stack.ServiceTemplateId))

	// When
	ctx = metadata.AppendToOutgoingContext(s.Context(), "x-device-id", uuid.NewString())
	resp2, err := tcc.GetPublicTelegramMiniappPromotion(ctx, &proto.GetPublicTelegramMiniappPromotionRequest{
		StackId: int32(promotion.StackID),
	})
	s.Require().NoError(err)
	s.Equal(int32(1), resp2.NumUniqueViews)

	// Then
	s.Equal("Test App", resp2.AppTitle)
	s.Require().NotNil(resp2.Stack)
	s.Equal(stack.ID, uint(resp2.Stack.Id))
	s.Require().NotNil(resp2.Stack.ServiceTemplateId)
	s.Equal(*stack.ServiceTemplateID, uint(*resp2.Stack.ServiceTemplateId))
}
