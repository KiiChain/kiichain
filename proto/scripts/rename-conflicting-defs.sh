#!/bin/sh
set -e

SWAGGER_DIR="./tmp-swagger-gen"

# List of conflicting definitions (space-separated)
CONFLICTING_DEFS="protobufAny runtimeError v1beta1Params v1beta1QueryParamsResponse v1beta1PageRequest v1QueryParamsResponse v1beta1Grant v1AccessType v1Params v1QueryCodeResponse v1beta1QueryDelegatorValidatorsResponse typesHeader v1Counterparty v1State"

echo "Renaming conflicting definitions to make them unique..."

# Find all swagger files and rename conflicting definitions
find "$SWAGGER_DIR" -name "*.swagger.json" | sort | while read file; do
  # Extract module path for creating unique names
  module_path=$(echo "$file" | sed 's|./tmp-swagger-gen/||' | sed 's|/query.swagger.json||' | sed 's|/service.swagger.json||')

  # Create a prefix from the module path (e.g., cosmos/bank/v1beta1 -> BankV1beta1)
  prefix=$(echo "$module_path" | awk -F'/' '{
    result = ""
    for (i = 2; i <= NF; i++) {
      # Capitalize first letter of each component
      word = $i
      result = result toupper(substr(word, 1, 1)) substr(word, 2)
    }
    print result
  }' | sed 's|-||g' | sed 's|_||g')

  modified=0

  for def in $CONFLICTING_DEFS; do
    # Check if this file has the definition
    if jq -e ".definitions.$def" "$file" > /dev/null 2>&1; then
      new_name="${prefix}${def}"
      # echo "  $file"
      # echo "    Renaming: $def -> $new_name"

      # Step 1: Rename the definition itself
      # Step 2: Update all $ref references to use the new name
      jq --arg old "$def" --arg new "$new_name" '
        # Rename the definition
        .definitions[$new] = .definitions[$old] |
        del(.definitions[$old]) |
        # Update all references in the entire document
        walk(
          if type == "object" and has("$ref") then
            if (.["$ref"] | test("#/definitions/" + $old + "$")) then
              .["$ref"] = "#/definitions/" + $new
            else
              .
            end
          else
            .
          end
        )
      ' "$file" > "${file}.tmp"

      mv "${file}.tmp" "$file"
      modified=1
    fi
  done
done

echo "✓ All conflicting definitions have been renamed"
