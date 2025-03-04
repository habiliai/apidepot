package prototest

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ApiDepotServerMock struct {
	proto.UnimplementedApiDepotServer
	mock.Mock
}

func (c *ApiDepotServerMock) InstallCustomVapi(ctx context.Context, request *proto.InstallCustomVapiRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) GetAllPublicTelegramMiniappPromotions(ctx context.Context, request *proto.GetAllPublicTelegramMiniappPromotionsRequest) (*proto.GetAllPublicTelegramMiniappPromotionsResponse, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.GetAllPublicTelegramMiniappPromotionsResponse), args.Error(1)
}

func (c *ApiDepotServerMock) CreateImageUploadUrl(ctx context.Context, request *proto.CreateImageUploadUrlRequest) (*proto.CreateImageUploadUrlResponse, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.CreateImageUploadUrlResponse), args.Error(1)
}

func (c *ApiDepotServerMock) UpdateTelegramMiniappPromotion(ctx context.Context, request *proto.UpdateTelegramMiniappPromotionRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) UninstallCustomVapi(ctx context.Context, request *proto.UninstallCustomVapiRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) GetCustomVapiByNameOnStack(ctx context.Context, request *proto.GetCustomVapiByNameOnStackRequest) (*proto.CustomVapi, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.CustomVapi), args.Error(1)
}

func (c *ApiDepotServerMock) UpdateCustomVapi(ctx context.Context, request *proto.UpdateCustomVapiRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) GetCustomVapisOnStack(ctx context.Context, id *proto.StackId) (*proto.GetCustomVapisOnStackResponse, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*proto.GetCustomVapisOnStackResponse), args.Error(1)
}

func (c *ApiDepotServerMock) GetPublicTelegramMiniappPromotion(ctx context.Context, request *proto.GetPublicTelegramMiniappPromotionRequest) (*proto.TelegramMiniappPromotion, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.TelegramMiniappPromotion), args.Error(1)
}

func (c *ApiDepotServerMock) GetUserStorageUsages(ctx context.Context, empty *emptypb.Empty) (*proto.GetUserStorageUsagesResponse, error) {
	args := c.Called(ctx, empty)
	return args.Get(0).(*proto.GetUserStorageUsagesResponse), args.Error(1)
}

func (c *ApiDepotServerMock) SetStackEnv(ctx context.Context, request *proto.SetStackEnvRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) UnsetStackEnv(ctx context.Context, request *proto.UnsetStackEnvRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) GetMyStorageUsage(ctx context.Context, empty *emptypb.Empty) (*proto.GetMyStorageUsageResponse, error) {
	args := c.Called(ctx, empty)
	return args.Get(0).(*proto.GetMyStorageUsageResponse), args.Error(1)
}

func (c *ApiDepotServerMock) HardForkGitRepo(ctx context.Context, request *proto.HardForkGitRepoRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) SearchServiceTemplates(ctx context.Context, request *proto.SearchServiceTemplatesRequest) (*proto.SearchServiceTemplatesResponse, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.SearchServiceTemplatesResponse), args.Error(1)
}

func (c *ApiDepotServerMock) RegisterCliApp(ctx context.Context, request *proto.RegisterCliAppRequest) (*proto.RegisterCliAppResponse, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.RegisterCliAppResponse), args.Error(1)
}

func (c *ApiDepotServerMock) DeleteCliApp(ctx context.Context, request *proto.DeleteCliAppRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) VerifyCliApp(ctx context.Context, request *proto.VerifyCliAppRequest) (*proto.VerifyCliAppResponse, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.VerifyCliAppResponse), args.Error(1)
}

func (c *ApiDepotServerMock) CreateStackFromServiceTemplate(ctx context.Context, request *proto.CreateStackFromServiceTemplateRequest) (*proto.Stack, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.Stack), args.Error(1)
}

func (c *ApiDepotServerMock) GetServiceTemplateById(ctx context.Context, id *proto.ServiceTemplateId) (*proto.ServiceTemplate, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*proto.ServiceTemplate), args.Error(1)
}

func (c *ApiDepotServerMock) CreateInstance(ctx context.Context, request *proto.CreateInstanceRequest) (*proto.Instance, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.Instance), args.Error(1)
}

func (c *ApiDepotServerMock) GetInstanceById(ctx context.Context, id *proto.InstanceId) (*proto.Instance, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*proto.Instance), args.Error(1)
}

func (c *ApiDepotServerMock) EditInstance(ctx context.Context, request *proto.EditInstanceRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) DeleteInstance(ctx context.Context, id *proto.InstanceId) (*emptypb.Empty, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) DeployStack(ctx context.Context, request *proto.DeployStackRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) LaunchInstance(ctx context.Context, id *proto.InstanceId) (*emptypb.Empty, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) StopInstance(ctx context.Context, id *proto.InstanceId) (*emptypb.Empty, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) GetStacks(ctx context.Context, request *proto.GetStacksRequest) (*proto.GetStacksResponse, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.GetStacksResponse), args.Error(1)
}

func (c *ApiDepotServerMock) CreateStack(ctx context.Context, request *proto.CreateStackRequest) (*proto.Stack, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.Stack), args.Error(1)
}

func (c *ApiDepotServerMock) GetStackById(ctx context.Context, id *proto.StackId) (*proto.Stack, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*proto.Stack), args.Error(1)
}

func (c *ApiDepotServerMock) DeleteStack(ctx context.Context, id *proto.StackId) (*emptypb.Empty, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) InstallAuth(ctx context.Context, request *proto.InstallAuthRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) UninstallAuth(ctx context.Context, id *proto.StackId) (*emptypb.Empty, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) InstallPostgrest(ctx context.Context, request *proto.InstallPostgrestRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) UninstallPostgrest(ctx context.Context, id *proto.StackId) (*emptypb.Empty, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) InstallStorage(ctx context.Context, request *proto.InstallStorageRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) UninstallStorage(ctx context.Context, id *proto.StackId) (*emptypb.Empty, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) InstallVapi(ctx context.Context, request *proto.InstallVapiRequest) (*proto.StackVapi, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.StackVapi), args.Error(1)
}

func (c *ApiDepotServerMock) UninstallVapi(ctx context.Context, request *proto.UninstallVapiRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) UpdateVapi(ctx context.Context, request *proto.UpdateVapiRequest) (*proto.StackVapi, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.StackVapi), args.Error(1)
}

func (c *ApiDepotServerMock) MigrateDatabase(ctx context.Context, request *proto.MigrateDatabaseRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) GetStackInstances(ctx context.Context, id *proto.StackId) (*proto.GetStackInstancesResponse, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*proto.GetStackInstancesResponse), args.Error(1)
}

func (c *ApiDepotServerMock) UpdateStack(ctx context.Context, request *proto.UpdateStackRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) GetProjects(ctx context.Context, request *proto.GetProjectsRequest) (*proto.GetProjectsResponse, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.GetProjectsResponse), args.Error(1)
}

func (c *ApiDepotServerMock) CreateProject(ctx context.Context, request *proto.CreateProjectRequest) (*proto.Project, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.Project), args.Error(1)
}

func (c *ApiDepotServerMock) GetProjectById(ctx context.Context, id *proto.ProjectId) (*proto.Project, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*proto.Project), args.Error(1)
}

func (c *ApiDepotServerMock) DeleteProject(ctx context.Context, id *proto.ProjectId) (*emptypb.Empty, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) UpdateVapiVersion(ctx context.Context, request *proto.UpdateVapiVersionRequest) (*proto.VapiReleaseId, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.VapiReleaseId), args.Error(1)
}

func (c *ApiDepotServerMock) UpdateGithubInstallationInfo(ctx context.Context, request *proto.UpdateUserGithubInstallationInfoRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) SyncExistingInstallation(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	args := c.Called(ctx, empty)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) GenerateInstallationAccessToken(ctx context.Context, empty *emptypb.Empty) (*proto.GenerateInstallationAccessTokenResponse, error) {
	args := c.Called(ctx, empty)
	return args.Get(0).(*proto.GenerateInstallationAccessTokenResponse), args.Error(1)
}

func (c *ApiDepotServerMock) UpdateUserProfile(ctx context.Context, request *proto.UpdateUserProfileRequest) (*emptypb.Empty, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) GetUser(ctx context.Context, empty *emptypb.Empty) (*proto.User, error) {
	args := c.Called(ctx, empty)
	return args.Get(0).(*proto.User), args.Error(1)
}

func (c *ApiDepotServerMock) GetVapiPackagesByOwnerId(ctx context.Context, id *proto.UserId) (*proto.GetVapiPackagesResponse, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*proto.GetVapiPackagesResponse), args.Error(1)
}

func (c *ApiDepotServerMock) RegisterVapi(ctx context.Context, request *proto.RegisterVapiRequest) (*proto.VapiReleaseId, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.VapiReleaseId), args.Error(1)
}

func (c *ApiDepotServerMock) GetVapiDocsUrl(ctx context.Context, request *proto.GetVapiDocsUrlRequest) (*proto.GetVapiDocsUrlResponse, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.GetVapiDocsUrlResponse), args.Error(1)
}

func (c *ApiDepotServerMock) GetUserByAuthId(ctx context.Context, id *proto.UserAuthId) (*proto.User, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*proto.User), args.Error(1)
}

func (c *ApiDepotServerMock) DeleteAllVapiReleases(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	args := c.Called(ctx, empty)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) DeleteAllVapiReleasesInPackage(ctx context.Context, id *proto.VapiPackageId) (*emptypb.Empty, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) GetVapiReleaseById(ctx context.Context, id *proto.VapiReleaseId) (*proto.VapiRelease, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*proto.VapiRelease), args.Error(1)
}

func (c *ApiDepotServerMock) DeleteVapiRelease(ctx context.Context, id *proto.VapiReleaseId) (*emptypb.Empty, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) DeleteVapiPackages(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	args := c.Called(ctx, empty)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) SearchVapis(ctx context.Context, request *proto.SearchVapisRequest) (*proto.SearchVapisResponse, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.SearchVapisResponse), args.Error(1)
}

func (c *ApiDepotServerMock) GetVapiReleasesInPackage(ctx context.Context, id *proto.VapiPackageId) (*proto.GetVapiReleasesResponse, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*proto.GetVapiReleasesResponse), args.Error(1)
}

func (c *ApiDepotServerMock) GetVapiReleaseByVersionInPackage(ctx context.Context, request *proto.GetVapiReleaseByVersionInPackageRequest) (*proto.VapiRelease, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.VapiRelease), args.Error(1)
}

func (c *ApiDepotServerMock) GetVapiPackages(ctx context.Context, request *proto.GetVapiPackagesRequest) (*proto.GetVapiPackagesResponse, error) {
	args := c.Called(ctx, request)
	return args.Get(0).(*proto.GetVapiPackagesResponse), args.Error(1)
}

func (c *ApiDepotServerMock) GetVapiPackageById(ctx context.Context, id *proto.VapiPackageId) (*proto.VapiPackage, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*proto.VapiPackage), args.Error(1)
}

func (c *ApiDepotServerMock) ResetSchema(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	args := c.Called(ctx, empty)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (c *ApiDepotServerMock) DeleteVapiPackage(ctx context.Context, id *proto.VapiPackageId) (*emptypb.Empty, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

var (
	_ proto.ApiDepotServer = (*ApiDepotServerMock)(nil)
)
