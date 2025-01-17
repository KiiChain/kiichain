# Upgrade tests

The main objective of this repository is to test upgrades and compatibilities between versions.

# How to use

Here you will find two integration tests:
- [Compatibility Test](integration_test/upgrade/run_compatibility_test.sh)
  - Used to check if different versions have consensus breaking versions
- [Upgrade Tests](integration_test/upgrade/run_upgrade_test.sh)
  - Used to run upgrade tests between versions
  - This can be used to check if upgrade snippets have any consensus breaking logic on node restart
  - THIS IS TODO

## Compatibility Test

The compatibility tests does the following:
1. Start the docker ecosystem using a old binary
2. Binaries are replaced in a partial manner:
- Only half the validators have the binary replaced
3. Check the status of the nodes
4. Replace the binaries on the other nodes before closing the tests

This forces a state were the node was upgraded with no proposal.

### How to run

Before running you must do some preparation:
- Since we will test the compatibility between versions you will need to pre-compile the old version
- The old version of the binary must be placed on:
  - integration_test/upgrade/old_binary

After this you just need to run:

```bash
integration_test/upgrade/run_compatibility_test.sh
```

## Upgrade Tests

TODO: Apply the docs after we have the scripts