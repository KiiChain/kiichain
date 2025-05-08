#!/usr/bin/env python3

import os
import sys
import shutil
import subprocess
import json


def check_solc_installed():
    if shutil.which("solc") is None:
        print("solc is not installed. Please install it to use this script.")
        sys.exit(1)


def run_solc(abi_path, output_dir):
    # Run solc to generate the ABI
    try:
        subprocess.run(
            ["solc", "--abi", "--overwrite", "-o", output_dir, abi_path], check=True
        )
    except subprocess.CalledProcessError as e:
        print(f"Error running solc: {e}")
        sys.exit(1)


def process_abi_hardhat(sol_file):
    # Get the output directory
    output_dir = os.path.dirname(sol_file)

    # Get the base name of the .sol file
    base_name = os.path.basename(sol_file).replace(".sol", "")

    # Load the ABI from the generated file
    abi_file = os.path.join(output_dir, f"{base_name}.abi")

    # Check if the ABI file exists
    if not os.path.exists(abi_file):
        print(f"ABI file {abi_file} does not exist.")
        return

    with open(abi_file, "r") as f:
        abi = json.load(f)

    # Create the hardhat format
    hardhat_abi = {
        "_format": "hh-sol-artifact-1",
        "contractName": base_name,
        "sourceName": sol_file,
        "abi": abi,
        "bytecode": "0x",
        "deployedBytecode": "0x",
        "linkReferences": {},
        "deployedLinkReferences": {},
    }

    # Delete the old ABI file and save the new one as abi.json
    os.remove(abi_file)
    with open(os.path.join(output_dir, "abi.json"), "w") as f:
        json.dump(hardhat_abi, f, indent=4)


def process_sol_files(abi_path):
    # Get the list of .sol files in the abi_path recursively
    sol_files = []
    for root, _, files in os.walk(abi_path):
        for file in files:
            if file.endswith(".sol"):
                sol_files.append(os.path.join(root, file))

    # Check if there are any .sol files
    if not sol_files:
        print(f"No .sol files found in {abi_path}")
        sys.exit(1)

    # Run solc for each .sol file
    for sol_file in sol_files:
        print(f"Processing {sol_file}...")

        # Get the output directory
        output_dir = os.path.dirname(sol_file)

        # Run solc to generate the ABI
        run_solc(sol_file, output_dir)

        # Process to the hardhat format
        process_abi_hardhat(sol_file)


if __name__ == "__main__":
    # Check if enough arguments were passed
    if len(sys.argv) != 2:
        print("Usage: solc_to_hardhat.py <abi_path>")
        sys.exit(1)

    # Get the abi path from the args
    abi_path = sys.argv[1]

    # Check if solc is installed
    check_solc_installed()

    # Process the sol files
    process_sol_files(abi_path)
