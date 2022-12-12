package entity

import "github.com/SevereCloud/vksdk/v2/object"

// BanReason describes ban reason.
type BanReason string

// Available ban reasons.
const (
	BanReasonNone           BanReason = "none"
	BanReasonPersonNonGrata BanReason = "person_non_grata"
)

// HeuristicRules describes heuristic rules.
type HeuristicRules struct {
	PersonNonGrata []HeuristicPersonNonGrataRule `toml:"person_non_grata"`
}

// Check checks if user qualifies for heuristics.
func (rr *HeuristicRules) Check(user *object.UsersUser) (bool, BanReason) {
	for _, rule := range rr.PersonNonGrata {
		if rule.Check(user) {
			return true, BanReasonPersonNonGrata
		}
	}

	return false, BanReasonNone
}

// HeuristicPersonNonGrataRule describes person non grata rule.
type HeuristicPersonNonGrataRule struct {
	Name      *string `toml:"name"`
	BirthDate *string `toml:"birth_date"`
}

// Check checks if user qualifies for rule.
func (r HeuristicPersonNonGrataRule) Check(user *object.UsersUser) bool {
	matches := 0

	if r.Name != nil {
		name := user.FirstName + " " + user.LastName
		if name == *r.Name {
			matches++
		}
	}
	if r.BirthDate != nil && user.Bdate == *r.BirthDate {
		matches++
	}

	return matches == r.assertCount()
}

func (r HeuristicPersonNonGrataRule) assertCount() int {
	count := 0

	if r.Name != nil {
		count++
	}
	if r.BirthDate != nil {
		count++
	}

	return count
}
