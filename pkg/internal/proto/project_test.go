package proto_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/proto"
	"github.com/stretchr/testify/mock"
)

func (s *ProtoTestSuite) TestGetProjects() {

	s.Run("when with no projects, should be get zero", func() {
		client, dispose := s.newClient()
		defer dispose()

		s.projects.On("GetProjects", mock.Anything, mock.Anything).Return([]domain.Project{}, nil).Once()
		defer s.projects.AssertExpectations(s.T())

		resp, err := client.GetProjects(s.Context(), &proto.GetProjectsRequest{})
		s.Require().NoError(err)

		projects := resp.Projects
		s.Require().Len(projects, 0)
	})

	s.Run("when with project once, should be get one", func() {
		client, dispose := s.newClient()
		defer dispose()

		s.projects.On("GetProjects", mock.Anything, mock.Anything).Return([]domain.Project{
			{
				Model: domain.Model{
					ID: 1,
				},
				Name: "test",
			},
		}, nil).Once()
		defer s.projects.AssertExpectations(s.T())

		res, err := client.GetProjects(s.Context(), &proto.GetProjectsRequest{})

		s.Require().NoError(err)
		s.Require().Len(res.Projects, 1)
		s.Require().Equal("test", res.Projects[0].Name)
	})
}

func (s *ProtoTestSuite) TestGetProject() {
	s.Run("when inserted a project with many stakcs, should be get a project with all stacks", func() {
		// Given
		project := domain.Project{
			Model: domain.Model{
				ID: 1,
			},
			Name: "test",
			Stacks: []domain.Stack{
				{
					Model: domain.Model{
						ID: 1,
					},
					Name:      "test1",
					Hash:      "hash1",
					ProjectID: 1,
				},
				{
					Model: domain.Model{
						ID: 2,
					},
					Name:      "test2",
					Hash:      "hash2",
					ProjectID: 1,
				},
			},
		}
		s.projects.On("GetProject", mock.Anything, uint(1)).Return(&project, nil).Once()
		defer s.projects.AssertExpectations(s.T())

		// When
		client, dispose := s.newClient()
		defer dispose()

		resp, err := client.GetProjectById(s.Context(), &proto.ProjectId{Id: 1})
		s.Require().NoError(err)

		// Then
		{
			s.Require().Equal(project.Name, resp.Name)
			s.Require().Len(resp.Stacks, 2)
			s.Require().Equal("test1", resp.Stacks[0].Name)
			s.Require().Equal("test2", resp.Stacks[1].Name)
		}
	})
}
