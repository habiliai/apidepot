package proto

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/mokiat/gog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *apiDepotServer) DeleteVapiPackage(ctx context.Context, id *VapiPackageId) (*emptypb.Empty, error) {
	if err := s.vapiService.DeletePackage(ctx, uint(id.GetId())); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) DeleteAllVapiReleasesInPackage(ctx context.Context, id *VapiPackageId) (*emptypb.Empty, error) {
	if err := s.vapiService.DeleteReleasesByPackageId(ctx, uint(id.GetId())); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) DeleteAllVapiReleases(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.vapiService.DeleteAllReleases(ctx); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) GetVapiReleaseById(ctx context.Context, id *VapiReleaseId) (*VapiRelease, error) {
	rel, err := s.vapiService.GetRelease(ctx, uint(id.GetId()))
	if err != nil {
		return nil, err
	}

	return newVapiReleasePbFromDb(rel), nil
}

func (s *apiDepotServer) DeleteVapiReleaseById(ctx context.Context, id *VapiReleaseId) (*emptypb.Empty, error) {
	if err := s.vapiService.DeleteRelease(ctx, uint(id.GetId())); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) RegisterVapi(ctx context.Context, request *RegisterVapiRequest) (*VapiReleaseId, error) {
	rel, err := s.vapiService.Register(ctx,
		request.GitRepo,
		request.GitBranch,
		request.Name,
		request.Description,
		request.Domains,
		request.VapiPoolId,
		request.Homepage,
	)
	if err != nil {
		return nil, err
	}

	return &VapiReleaseId{
		Id: int32(rel.ID),
	}, nil
}

func (s *apiDepotServer) SearchVapis(ctx context.Context, request *SearchVapisRequest) (*SearchVapisResponse, error) {
	output, err := s.vapiService.SearchVapis(ctx, vapi.SearchVapisInput{
		Name:     request.Name,
		Version:  request.Version,
		PageNum:  int(request.PageNum),
		PageSize: int(request.PageSize),
	})
	if err != nil {
		return nil, err
	}

	var response SearchVapisResponse
	if output.NextPage != nil {
		response.NextPage = gog.PtrOf(int32(*output.NextPage))
	}

	for _, rel := range output.Releases {
		relPb := newVapiReleasePbFromDb(&rel)
		response.Releases = append(response.Releases, relPb)
	}

	response.NumTotal = int32(output.NumTotal)

	return &response, nil
}

func (s *apiDepotServer) DeleteVapiRelease(ctx context.Context, id *VapiReleaseId) (*emptypb.Empty, error) {
	if err := s.vapiService.DeleteRelease(ctx, uint(id.GetId())); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) DeleteVapiPackages(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.vapiService.DeleteAllPackages(ctx, 0); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *apiDepotServer) GetVapiReleasesInPackage(ctx context.Context, id *VapiPackageId) (*GetVapiReleasesResponse, error) {
	pkg, err := s.vapiService.GetPackage(ctx, uint(id.Id))
	if err != nil {
		return nil, err
	}

	var resp GetVapiReleasesResponse
	for _, rel := range pkg.Releases {
		relPb := newVapiReleasePbFromDb(&rel)
		resp.Releases = append(resp.Releases, relPb)
	}

	return &resp, nil
}

func (s *apiDepotServer) GetVapiReleaseByVersionInPackage(
	ctx context.Context,
	req *GetVapiReleaseByVersionInPackageRequest,
) (*VapiRelease, error) {
	rel, err := s.vapiService.GetReleaseByVersionInPackage(ctx, uint(req.PackageId), req.Version)
	if err != nil {
		return nil, err
	}

	return newVapiReleasePbFromDb(rel), nil
}

func (s *apiDepotServer) GetVapiPackages(ctx context.Context, req *GetVapiPackagesRequest) (*GetVapiPackagesResponse, error) {
	packages, err := s.vapiService.GetPackages(ctx, vapi.GetPackagesInput{
		Name: req.Name,
	})
	if err != nil {
		return nil, err
	}

	var response GetVapiPackagesResponse
	for _, pkg := range packages {
		pkgPb := newVapiPackagePbFromDb(&pkg)
		response.Packages = append(response.Packages, pkgPb)
	}

	return &response, nil
}

func (s *apiDepotServer) GetVapiPackageById(ctx context.Context, id *VapiPackageId) (*VapiPackage, error) {
	pkg, err := s.vapiService.GetPackage(ctx, uint(id.GetId()))
	if err != nil {
		return nil, err
	}

	return newVapiPackagePbFromDb(pkg), nil
}

func (s *apiDepotServer) GetVapiDocsUrl(ctx context.Context, request *GetVapiDocsUrlRequest) (*GetVapiDocsUrlResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVapiDocsUrl not implemented")
}

func (s *apiDepotServer) GetVapiPackagesByOwnerId(ctx context.Context, request *UserId) (*GetVapiPackagesResponse, error) {
	packages, err := s.vapiService.GetPackagesByOwnerId(ctx, uint(request.GetId()))
	if err != nil {
		return nil, err
	}

	var response GetVapiPackagesResponse
	for _, pkg := range packages {
		pkgPb := newVapiPackagePbFromDb(&pkg)
		response.Packages = append(response.Packages, pkgPb)
	}

	return &response, nil
}

func (s *apiDepotServer) UpdateVapiVersion(ctx context.Context, req *UpdateVapiVersionRequest) (*VapiReleaseId, error) {
	tx := helpers.GetTx(ctx)
	pkg, err := domain.FindVapiPackageByID(tx, uint(req.GetPackageId()))
	if err != nil {
		return nil, err
	}

	rel, err := s.vapiService.Register(
		ctx,
		pkg.GitRepo,
		pkg.GitBranch,
		pkg.Name,
		req.Description,
		req.Domains,
		pkg.VapiPoolId,
		req.Homepage,
	)
	if err != nil {
		return nil, err
	}

	return &VapiReleaseId{Id: int32(rel.ID)}, nil
}
