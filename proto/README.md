# Code generation

To generate the code for the protobuf files, first install the `ignite` tool.
We need version v0.23.0, which is outdated, but works with the current version of the codebase.
Pull binaries from the [releases page](https://github.com/ignite/cli/releases/tag/v0.23.0) or install from source code 
following instructions.

Verify the installation by running `ignite version`:

```bash
% ignite version          
·
· 🛸 Ignite CLI v28.2.0 is available!
·
· To upgrade your Ignite CLI version, see the upgrade doc: https://docs.ignite.com/guide/install.html#upgrading-your-ignite-cli-installation
·
··

Ignite CLI version:     v0.23.0
Ignite CLI build date:  2022-07-24T18:17:44Z
Ignite CLI source hash: 64df9aef958b3e8bc04b40d9feeb03426075ea89
...More information...
```

Then, to generate the code, run the following command:

```bash
ignite generate proto-go
```