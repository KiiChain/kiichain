package types

import (
	"fmt"

	"gopkg.in/yaml.v2"

	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v1/x/oracle/utils"
)

// Default parameter value
var (
	DefaultVotePeriod    = uint64(2)                         // Voting every two blocks
	DefaultSlashWindow   = utils.BlocksPerDay * 2            // 2 days for oracle slashing
	DefaultVoteThreshold = math.LegacyNewDecWithPrec(667, 3) // 0.667 | 66.7%
	DefaultRewardBand    = math.LegacyNewDecWithPrec(2, 2)   // 0.02% | 2%
	DefaultWhitelist     = DenomList{
		{Name: utils.MicroBtcDenom},
		{Name: utils.MicroEthDenom},
		{Name: utils.MicroSolDenom},
		{Name: utils.MicroXrpDenom},
		{Name: utils.MicroBnbDenom},
		{Name: utils.MicroUsdtDenom},
		{Name: utils.MicroUsdcDenom},
	}
	DefaultSlashFraction     = math.LegacyNewDecWithPrec(0, 4) // 0.00 | 0%
	DefaultMinValidPerWindow = math.LegacyNewDecWithPrec(5, 2) // 0.05 | 5%
	DefaultLookbackDuration  = uint64(3600)
)

// DefaultParams returns the default oracle module parameters
func DefaultParams() Params {
	return Params{
		VotePeriod:        DefaultVotePeriod,
		VoteThreshold:     DefaultVoteThreshold,
		RewardBand:        DefaultRewardBand,
		Whitelist:         DefaultWhitelist,
		SlashFraction:     DefaultSlashFraction,
		SlashWindow:       DefaultSlashWindow,
		MinValidPerWindow: DefaultMinValidPerWindow,
		LookbackDuration:  DefaultLookbackDuration,
	}
}

// String implements fmt.Stringer interface. Format the parameters as Yaml and return as string
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate performs basic validation on oracle parameters.
func (p Params) Validate() error {
	if p.VotePeriod == 0 {
		return fmt.Errorf("oracle parameter VotePeriod must be > 0, is %d", p.VotePeriod)
	}
	if p.VoteThreshold.LTE(math.LegacyNewDecWithPrec(33, 2)) {
		return fmt.Errorf("oracle parameter VoteThreshold must be greater than 33 percent")
	}

	if p.RewardBand.GT(math.LegacyOneDec()) || p.RewardBand.IsNegative() {
		return fmt.Errorf("oracle parameter RewardBand must be between [0, 1]")
	}

	if p.SlashFraction.GT(math.LegacyOneDec()) || p.SlashFraction.IsNegative() {
		return fmt.Errorf("oracle parameter SlashFraction must be between [0, 1]")
	}

	if p.SlashWindow < p.VotePeriod {
		return fmt.Errorf("oracle parameter SlashWindow must be greater than or equal with VotePeriod")
	}

	if p.SlashWindow%p.VotePeriod != 0 {
		return fmt.Errorf("oracle parameter SlashWindow must be divisible by VotePeriod")
	}

	if p.MinValidPerWindow.GT(math.LegacyOneDec()) || p.MinValidPerWindow.IsNegative() {
		return fmt.Errorf("oracle parameter MinValidPerWindow must be between [0, 1]")
	}

	for _, denom := range p.Whitelist {
		if len(denom.Name) == 0 {
			return fmt.Errorf("oracle parameter Whitelist Denom must have name")
		}
	}
	return nil
}

func validateVotePeriod(i interface{}) error {
	v, ok := i.(uint64) // Data type must be uint64
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("vote period must be positive: %d", v)
	}
	return nil
}

func validateVoteThreshold(i interface{}) error {
	v, ok := i.(math.LegacyDec) // Data type must be Decimal from cosmos sdk
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("vote threshold must be bigger than 0%%: %s", v)
	}

	if v.GT(math.LegacyOneDec()) { // Parameter cannot be greater than 1.00
		return fmt.Errorf("vote threshold too large: %s", v)
	}

	return nil
}

func validateRewardBand(i interface{}) error {
	v, ok := i.(math.LegacyDec) // Data type must be Decimal from cosmos sdk
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("reward band must be positive: %s", v)
	}

	if v.GT(math.LegacyOneDec()) { // Parameter cannot be greater than 1.00
		return fmt.Errorf("reward band is too large: %s", v)
	}

	return nil
}

func validateWhitelist(i interface{}) error {
	v, ok := i.(DenomList) // Data type must be DenomList
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, denom := range v {
		if len(denom.Name) == 0 {
			return fmt.Errorf("oracle parameter Whitelist Denom must have elements")
		}
	}

	return nil
}

func validateSlashFraction(i interface{}) error {
	v, ok := i.(math.LegacyDec) // Data type must be Decimal from cosmos sdk
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("slash fraction must be positive: %s", v)
	}

	if v.GT(math.LegacyOneDec()) { // Parameter cannot be greater than 1.00
		return fmt.Errorf("slash fraction is too large: %s", v)
	}

	return nil
}

func validateSlashWindow(i interface{}) error {
	v, ok := i.(uint64) // Data type must be uint64
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("slash window must be positive: %d", v)
	}

	return nil
}

func validateMinValidPerWindow(i interface{}) error {
	v, ok := i.(math.LegacyDec) // Data type must be Decimal from cosmos sdk
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min valid per window must be positive: %s", v)
	}

	if v.GT(math.LegacyOneDec()) { // Parameter cannot be greater than 1.00
		return fmt.Errorf("min valid per window is too large: %s", v)
	}

	return nil
}

func validateLookbackDuration(i interface{}) error {
	_, ok := i.(uint64) // Data type must be uint64
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
