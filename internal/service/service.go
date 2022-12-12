package service

import (
	"fmt"
	"log"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"

	"github.com/sklyar/vk-banhammer/internal/entity"
)

// VkClient is a VK API client.
// It is used to mock VK API client in tests.
//
//go:generate mockgen -source=service.go -package=main -destination=service_mock.go VkClient
type VkClient interface {
	UsersGet(params api.Params) (api.UsersGetResponse, error)
	GroupsBan(params api.Params) (int, error)
	WallDeleteComment(params api.Params) (int, error)
}

// Service is a banhammer service.
type Service struct {
	heuristicRules entity.HeuristicRules
	client         VkClient
}

// NewService creates a new banhammer service.
func NewService(client VkClient, heuristicRules entity.HeuristicRules) *Service {
	return &Service{heuristicRules: heuristicRules, client: client}
}

// CheckComment checks comment and ban user if needed.
func (s *Service) CheckComment(comment *entity.Comment) (entity.BanReason, error) {
	user, err := s.getUserByID(comment.FromID)
	if err != nil {
		return entity.BanReasonNone, fmt.Errorf("failed to get user: %w", err)
	}

	log.Println("user:", user.FirstName, user.LastName, user.Bdate)
	banned, reason := s.heuristicRules.Check(user)
	if banned {
		if err := s.banUser(comment.OwnerID, user.ID, reason); err != nil {
			return reason, err
		}

		if err := s.deleteComment(comment); err != nil {
			return reason, err
		}
	}

	return reason, nil
}

func (s *Service) getUserByID(userID int) (*object.UsersUser, error) {
	users, err := s.client.UsersGet(
		api.Params{
			"user_ids": userID,
			"fields":   "bdate",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &users[0], nil
}

func (s *Service) banUser(groupID, userID int, reason entity.BanReason) error {
	resp, err := s.client.GroupsBan(
		api.Params{
			"group_id":        -groupID, // group id should be negative
			"owner_id":        userID,
			"comment":         string(reason),
			"comment_visible": 0,
		},
	)
	if err != nil {
		return err
	}

	if resp != 1 {
		return fmt.Errorf("response is not success")
	}

	return nil
}

func (s *Service) deleteComment(comment *entity.Comment) error {
	result, err := s.client.WallDeleteComment(api.Params{
		"owner_id":   comment.OwnerID,
		"comment_id": comment.ID,
	})
	if err != nil {
		return err
	}

	if result != 1 {
		return fmt.Errorf("response is not success")
	}

	return nil
}
