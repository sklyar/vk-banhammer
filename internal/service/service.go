package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/sklyar/vk-banhammer/internal/entity"
	"go.uber.org/zap"
)

const cacheSize = 1000

var (
	// ErrUserNotFound is returned when user is not found.
	ErrUserNotFound = errors.New("user not found")

	// ErrBadResponse is returned when VK API returns bad response.
	ErrBadResponse = errors.New("bad response")
)

// VkClient is a VK API client.
// It is used to mock VK API client in tests.
//
//go:generate mockgen -source=service.go -package=service -destination=service_mock.go VkClient
type VkClient interface {
	UsersGet(params api.Params) (api.UsersGetResponse, error)
	GroupsBan(params api.Params) (int, error)
	WallDeleteComment(params api.Params) (int, error)
}

// Service is a banhammer service.
type Service struct {
	heuristicRules entity.HeuristicRules
	client         VkClient

	cache *lru.Cache[int, *object.UsersUser]
	m     sync.RWMutex

	logger *zap.Logger
}

// NewService creates a new banhammer service.
func NewService(logger *zap.Logger, client VkClient, heuristicRules entity.HeuristicRules) *Service {
	cache, err := lru.New[int, *object.UsersUser](cacheSize)
	if err != nil {
		panic(err)
	}

	return &Service{
		heuristicRules: heuristicRules,
		client:         client,
		cache:          cache,
		m:              sync.RWMutex{},
		logger:         logger,
	}
}

// CheckComment checks comment and ban user if needed.
func (s *Service) CheckComment(comment *entity.Comment) (entity.BanReason, error) {
	user, err := s.getUserByID(comment.FromID)
	if err != nil {
		// Ignore comments from groups.
		if isCommentFromGroup(comment) {
			return entity.BanReasonNone, nil
		}
		s.logger.Error("failed to get user", zap.Error(err), zap.Reflect("comment", comment))
		return entity.BanReasonNone, fmt.Errorf("failed to get user: %w", err)
	}

	s.logger.Debug(
		"new comment",
		zap.Int("id", user.ID),
		zap.String("first_name", user.FirstName),
		zap.String("last_name", user.LastName),
		zap.String("bday", user.Bdate),
	)

	reason, shouldBan := s.heuristicRules.Check(user)
	if shouldBan {
		if err := s.banUser(comment.OwnerID, user.ID, reason); err != nil {
			return reason, fmt.Errorf("failed to ban user: %w", err)
		}

		if err := s.deleteComment(comment); err != nil {
			return reason, fmt.Errorf("failed to delete comment: %w", err)
		}
	}

	return reason, nil
}

func (s *Service) getUserByID(userID int) (*object.UsersUser, error) {
	u, exists := s.cache.Get(userID)
	if exists {
		return u, nil
	}

	users, err := s.client.UsersGet(
		api.Params{
			"user_ids": userID,
			"fields":   "bdate",
		},
	)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, ErrUserNotFound
	}

	u = &users[0]
	s.cache.Add(userID, u)

	return u, nil
}

func (s *Service) banUser(groupID, userID int, reason entity.BanReason) error {
	req := api.Params{
		"group_id":        -groupID, // group id should be negative.
		"owner_id":        userID,
		"comment":         string(reason),
		"comment_visible": 0,
	}
	return s.do(s.client.GroupsBan, req)
}

func (s *Service) deleteComment(comment *entity.Comment) error {
	req := api.Params{
		"owner_id":   comment.OwnerID,
		"comment_id": comment.ID,
	}
	return s.do(s.client.WallDeleteComment, req)
}

func (s *Service) do(fn func(api.Params) (int, error), params api.Params) error {
	res, err := fn(params)
	if err != nil {
		return err
	}

	if res != 1 {
		return ErrBadResponse
	}

	return nil
}

func isCommentFromGroup(comment *entity.Comment) bool {
	return strings.HasPrefix(comment.Text, "[club") || comment.FromID < 0
}
