package service

import (
	"errors"
	"testing"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/golang/mock/gomock"

	"github.com/sklyar/vk-banhammer/internal/entity"
)

func TestServiceCheckComment(t *testing.T) {
	t.Parallel()

	defaultComment := &entity.Comment{
		ID:      1,
		FromID:  87524863,
		Date:    1580000000,
		Text:    "test",
		PostID:  911,
		OwnerID: -61061413,
	}

	type dependencies struct {
		client *MockVkClient
	}

	tests := []struct {
		name           string
		comment        *entity.Comment
		heuristicRules entity.HeuristicRules
		setup          func(*dependencies)
		want           entity.BanReason
		wantErr        bool
	}{
		{
			name:           "get user error",
			comment:        defaultComment,
			heuristicRules: entity.HeuristicRules{},
			setup: func(d *dependencies) {
				d.client.EXPECT().
					UsersGet(api.Params{
						"user_ids": 87524863,
						"fields":   "bdate",
					}).
					Return(api.UsersGetResponse{}, errors.New("some error"))
			},
			want:    entity.BanReasonNone,
			wantErr: true,
		},
		{
			name:           "user not found",
			comment:        defaultComment,
			heuristicRules: entity.HeuristicRules{},
			setup: func(d *dependencies) {
				d.client.EXPECT().
					UsersGet(api.Params{
						"user_ids": 87524863,
						"fields":   "bdate",
					}).
					Return(api.UsersGetResponse{}, nil)
			},
			want:    entity.BanReasonNone,
			wantErr: true,
		},
		{
			name:    "user not banned",
			comment: defaultComment,
			heuristicRules: entity.HeuristicRules{
				PersonNonGrata: []entity.HeuristicPersonNonGrataRule{
					{Name: toPtr("test")},
				},
			},
			setup: func(d *dependencies) {
				d.client.EXPECT().
					UsersGet(api.Params{
						"user_ids": 87524863,
						"fields":   "bdate",
					}).
					Return(
						[]object.UsersUser{
							{
								ID:        87524863,
								FirstName: "Bob",
								LastName:  "Marley",
							},
						},
						nil,
					)
			},
			want:    entity.BanReasonNone,
			wantErr: false,
		},
		{
			name:    "user banned by name",
			comment: defaultComment,
			heuristicRules: entity.HeuristicRules{
				PersonNonGrata: []entity.HeuristicPersonNonGrataRule{
					{Name: toPtr("Bob Marley")},
				},
			},
			setup: func(d *dependencies) {
				d.client.EXPECT().
					UsersGet(api.Params{
						"user_ids": 87524863,
						"fields":   "bdate",
					}).
					Return(
						[]object.UsersUser{
							{
								ID:        87524863,
								FirstName: "Bob",
								LastName:  "Marley",
							},
						},
						nil,
					)

				d.client.EXPECT().GroupsBan(api.Params{
					"group_id":        61061413,
					"owner_id":        87524863,
					"comment":         string(entity.BanReasonPersonNonGrata),
					"comment_visible": 0,
				}).Return(1, nil)

				d.client.EXPECT().WallDeleteComment(api.Params{
					"owner_id":   -61061413,
					"comment_id": 1,
				}).Return(1, nil)

			},
			want:    entity.BanReasonPersonNonGrata,
			wantErr: false,
		},
		{
			name:    "user banned by name and birthday",
			comment: defaultComment,
			heuristicRules: entity.HeuristicRules{
				PersonNonGrata: []entity.HeuristicPersonNonGrataRule{
					{Name: toPtr("Bob Marley"), BirthDate: toPtr("1.1.2000")},
				},
			},
			setup: func(d *dependencies) {
				d.client.EXPECT().
					UsersGet(api.Params{
						"user_ids": 87524863,
						"fields":   "bdate",
					}).
					Return(
						[]object.UsersUser{
							{
								ID:        87524863,
								FirstName: "Bob",
								LastName:  "Marley",
								Bdate:     "1.1.2000",
							},
						},
						nil,
					)

				d.client.EXPECT().GroupsBan(api.Params{
					"group_id":        61061413,
					"owner_id":        87524863,
					"comment":         string(entity.BanReasonPersonNonGrata),
					"comment_visible": 0,
				}).Return(1, nil)

				d.client.EXPECT().WallDeleteComment(api.Params{
					"owner_id":   -61061413,
					"comment_id": 1,
				}).Return(1, nil)
			},
			want:    entity.BanReasonPersonNonGrata,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			deps := dependencies{
				client: NewMockVkClient(ctrl),
			}
			if tt.setup != nil {
				tt.setup(&deps)
			}

			s := NewService(deps.client, tt.heuristicRules)
			got, err := s.CheckComment(tt.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckComment() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func toPtr[T any](v T) *T {
	return &v
}
