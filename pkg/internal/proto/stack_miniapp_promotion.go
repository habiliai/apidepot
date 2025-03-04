package proto

import (
	"context"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/util"
	"github.com/mokiat/gog"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/url"
	"strconv"
	"time"
)

func (s *apiDepotServer) UpdateTelegramMiniappPromotion(
	ctx context.Context,
	req *UpdateTelegramMiniappPromotionRequest,
) (*emptypb.Empty, error) {
	tx := helpers.GetTx(ctx)
	st, err := s.stackService.GetStack(ctx, uint(req.StackId))
	if err != nil {
		return nil, err
	}

	tmp := st.TelegramMiniappPromotion
	if tmp == nil {
		tmp = &domain.TelegramMiniappPromotion{}
	}
	tmp.StackID = st.ID
	tmp.Link = req.LinkUrl
	tmp.AppScreenshotImageUrls = req.AppScreenshotImageUrls
	tmp.Public = req.Public
	tmp.AppTitle = req.AppTitle
	tmp.AppDescription = req.AppDescription
	tmp.AppIconImageUrl = req.AppIconImageUrl
	tmp.AppBannerImageUrl = req.AppBannerImageUrl

	if err := tx.Transaction(func(tx *gorm.DB) error {
		return tmp.Save(tx)
	}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) GetPublicTelegramMiniappPromotion(
	ctx context.Context,
	req *GetPublicTelegramMiniappPromotionRequest,
) (*TelegramMiniappPromotion, error) {
	tx := helpers.GetTx(ctx)
	deviceId := helpers.GetDeviceId(ctx)

	if deviceId == "" {
		return nil, errors.Wrapf(tclerrors.ErrBadRequest, "x-device-id is required")
	}

	var tmp domain.TelegramMiniappPromotion
	if err := tx.
		Preload("Views").
		Preload("Stack").
		First(&tmp, "public = ? AND stack_id = ?", true, req.StackId).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find telegram miniapp promotion")
	}

	if err := tx.Transaction(func(tx *gorm.DB) error {
		return errors.Wrapf(tx.Clauses(
			clause.OnConflict{
				Columns: []clause.Column{{Name: "telegram_miniapp_promotion_id"}, {Name: "device_id"}},
				DoUpdates: clause.Set{
					{
						Column: clause.Column{Name: "num_views"},
						Value:  gorm.Expr("telegram_miniapp_promotion_views.num_views + ?", 1),
					},
				},
			},
		).Create(&domain.TelegramMiniappPromotionView{
			DeviceID:                   deviceId,
			TelegramMiniappPromotionID: tmp.ID,
			NumViews:                   1,
		}).Error, "failed to create telegram miniapp promotion view")
	}); err != nil {
		return nil, err
	}

	return newTelegramMiniappPromotionPbFromDb(tmp, true), nil
}

func (s *apiDepotServer) GetAllPublicTelegramMiniappPromotions(
	ctx context.Context,
	req *GetAllPublicTelegramMiniappPromotionsRequest,
) (*GetAllPublicTelegramMiniappPromotionsResponse, error) {
	tx := helpers.GetTx(ctx)

	stmt := tx.
		Model(&domain.TelegramMiniappPromotion{Public: true}).
		Joins(
			"LEFT JOIN (?) as views ON views.telegram_miniapp_promotion_id = telegram_miniapp_promotions.id",
			tx.Model(&domain.TelegramMiniappPromotionView{}).Select("telegram_miniapp_promotion_id, COUNT(*) as num_views").Group("telegram_miniapp_promotion_id"),
		).
		Select("telegram_miniapp_promotions.*, views.num_views")

	var numTotal int64
	if err := stmt.Count(&numTotal).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to count")
	}

	orderBy := GetAllPublicTelegramMiniappPromotionsRequest_ORDER_BY_NUM_VIEWS
	if req.OrderBy != nil {
		orderBy = *req.OrderBy
	}

	cursor := ""
	if req.Cursor != nil {
		cursor = *req.Cursor
	}
	cursorValues, err := url.ParseQuery(cursor)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse cursor")
	}

	switch orderBy {
	case GetAllPublicTelegramMiniappPromotionsRequest_ORDER_BY_NUM_VIEWS:
		{
			lastNumViews, err := strconv.Atoi(util.DefaultIfEmpty(cursorValues.Get("lastNumViews"), "-1"))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse lastNumViews")
			}
			lastId, err := strconv.Atoi(util.DefaultIfEmpty(cursorValues.Get("lastId"), "0"))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse lastId")
			}

			if lastNumViews >= 0 && lastId > 0 {
				stmt = stmt.Where("num_views < ? OR (num_views = ? AND id > ?)", lastNumViews, lastNumViews, lastId)
			}

			stmt = stmt.Order("num_views DESC, id ASC")
		}
	case GetAllPublicTelegramMiniappPromotionsRequest_ORDER_BY_CREATED_AT:
		{
			lastCreatedAt, err := strconv.ParseInt(util.DefaultIfEmpty(cursorValues.Get("lastCreatedAt"), "0"), 10, 64)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse lastCreatedAt")
			}
			lastId, err := strconv.Atoi(util.DefaultIfEmpty(cursorValues.Get("lastId"), "0"))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse lastId")
			}

			if lastCreatedAt > 0 && lastId > 0 {
				lastCreatedAt := time.UnixMicro(lastCreatedAt)
				stmt = stmt.Where("created_at < ? OR (created_at = ? AND id > ?)", lastCreatedAt, lastCreatedAt, lastId)
			}

			stmt = stmt.Order("created_at DESC, id ASC")
		}
	}

	limit := 10
	if req.Limit != nil {
		limit = int(*req.Limit)
	}

	type Record struct {
		domain.TelegramMiniappPromotion
		NumViews int
	}
	var records []Record
	if err := stmt.
		Preload("Stack").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to find")
	}

	nextCursor := ""
	if len(records) > 0 {
		lastRecord := records[len(records)-1]
		lastCreatedAt := lastRecord.CreatedAt.UnixMicro()
		nextCursorValues := url.Values{
			"lastId": []string{strconv.Itoa(int(lastRecord.ID))},
		}

		switch orderBy {
		case GetAllPublicTelegramMiniappPromotionsRequest_ORDER_BY_NUM_VIEWS:
			nextCursorValues["lastNumViews"] = []string{strconv.Itoa(lastRecord.NumViews)}
		case GetAllPublicTelegramMiniappPromotionsRequest_ORDER_BY_CREATED_AT:
			nextCursorValues["lastCreatedAt"] = []string{strconv.FormatInt(lastCreatedAt, 10)}
		}

		nextCursor = nextCursorValues.Encode()
	}

	return &GetAllPublicTelegramMiniappPromotionsResponse{
		Records: gog.Map(records, func(r Record) *TelegramMiniappPromotion {
			result := newTelegramMiniappPromotionPbFromDb(r.TelegramMiniappPromotion, true)
			result.NumUniqueViews = int32(r.NumViews)

			return result
		}),
		NumTotal:   int32(numTotal),
		NextCursor: nextCursor,
	}, nil
}
