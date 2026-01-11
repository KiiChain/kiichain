# Custom Invariant Checker for Rewards Module

## Overview
This package implements custom invariant checks for the rewards module, 
following Cosmos SDK v0.53+ patterns where the SDK no longer provides 
automatic invariant execution.

## Invariants Implemented

### 1. Release Schedule Consistency
- Released amount ≤ Total amount
- Active schedules have valid time ranges
- Denom consistency between released and total

### 2. Community Pool Non-Negative
- Community pool amounts are never negative
- Valid denom and amount values

### 3. Parameters Validation
- Token denom is not empty

## Architecture
- `invariants/types.go`: Interface to break circular dependency
- `invariants/*.go`: Individual invariant implementations
- `invariants.go`: Registration and public interface

## Usage
Invariants are registered via `RegisterInvariants` method in AppModule,
but require custom execution logic (not automatically run by SDK).

## Future Enhancements
1. Add custom invariant executor in EndBlocker
2. Configurable check intervals via app.toml
3. Metrics and alerting for violations
