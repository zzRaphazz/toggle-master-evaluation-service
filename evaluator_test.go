package main

import "testing"

func TestRunEvaluationLogic(t *testing.T) {
	tests := []struct {
		name string
		info *CombinedFlagInfo
		want bool
	}{
		{
			name: "missing flag is disabled",
			info: &CombinedFlagInfo{},
			want: false,
		},
		{
			name: "disabled flag stays disabled",
			info: &CombinedFlagInfo{Flag: &Flag{Name: "new-ui", IsEnabled: false}},
			want: false,
		},
		{
			name: "enabled flag without a rule is enabled for everyone",
			info: &CombinedFlagInfo{Flag: &Flag{Name: "new-ui", IsEnabled: true}},
			want: true,
		},
		{
			name: "disabled rule is treated as no restriction",
			info: &CombinedFlagInfo{
				Flag: &Flag{Name: "new-ui", IsEnabled: true},
				Rule: &TargetingRule{IsEnabled: false},
			},
			want: true,
		},
		{
			name: "unsupported rule type is disabled",
			info: &CombinedFlagInfo{
				Flag: &Flag{Name: "new-ui", IsEnabled: true},
				Rule: &TargetingRule{
					IsEnabled: true,
					Rules:     Rule{Type: "UNKNOWN", Value: 100.0},
				},
			},
			want: false,
		},
		{
			name: "non-numeric percentage is disabled",
			info: &CombinedFlagInfo{
				Flag: &Flag{Name: "new-ui", IsEnabled: true},
				Rule: &TargetingRule{
					IsEnabled: true,
					Rules:     Rule{Type: "PERCENTAGE", Value: "50"},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (&App{}).runEvaluationLogic(tt.info, "user-123"); got != tt.want {
				t.Fatalf("runEvaluationLogic() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestRunEvaluationLogicPercentageBoundary(t *testing.T) {
	const userID = "user-123"
	const flagName = "new-ui"
	bucket := getDeterministicBucket(userID + flagName)

	info := &CombinedFlagInfo{
		Flag: &Flag{Name: flagName, IsEnabled: true},
		Rule: &TargetingRule{
			IsEnabled: true,
			Rules:     Rule{Type: "PERCENTAGE", Value: float64(bucket)},
		},
	}
	if (&App{}).runEvaluationLogic(info, userID) {
		t.Fatalf("user bucket %d should not pass a percentage equal to the bucket", bucket)
	}

	info.Rule.Rules.Value = float64(bucket + 1)
	if !(&App{}).runEvaluationLogic(info, userID) {
		t.Fatalf("user bucket %d should pass a percentage of %d", bucket, bucket+1)
	}
}

func TestGetDeterministicBucket(t *testing.T) {
	const input = "user-123new-ui"

	first := getDeterministicBucket(input)
	if second := getDeterministicBucket(input); first != second {
		t.Fatalf("getDeterministicBucket() = %d and %d for the same input", first, second)
	}
	if first < 0 || first > 99 {
		t.Fatalf("getDeterministicBucket() = %d, want a value in [0, 99]", first)
	}
}
