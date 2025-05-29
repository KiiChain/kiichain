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
