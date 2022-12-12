package main

import (
	"testing"

	"github.com/sklyar/vk-banhammer/internal/entity"
)

func TestValidateHeuristicRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rules   entity.HeuristicRules
		wantErr bool
	}{
		{
			name: "valid person non grata rule with name",
			rules: entity.HeuristicRules{
				PersonNonGrata: []entity.HeuristicPersonNonGrataRule{
					{Name: toPtr("test")},
				},
			},
			wantErr: false,
		},
		{
			name: "valid person non grata rule with birthday",
			rules: entity.HeuristicRules{
				PersonNonGrata: []entity.HeuristicPersonNonGrataRule{
					{BirthDate: toPtr("1.11.2000")},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid format birthday",
			rules: entity.HeuristicRules{
				PersonNonGrata: []entity.HeuristicPersonNonGrataRule{
					{BirthDate: toPtr("01.11.2000")},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid value birthday",
			rules: entity.HeuristicRules{
				PersonNonGrata: []entity.HeuristicPersonNonGrataRule{
					{BirthDate: toPtr("1.13.2000")},
				},
			},
			wantErr: true,
		},
		{
			name:    "empty rules",
			rules:   entity.HeuristicRules{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := validateHeuristicRules(tt.rules); (err != nil) != tt.wantErr {
				t.Errorf("validateHeuristicRules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func toPtr[T any](v T) *T {
	return &v
}
