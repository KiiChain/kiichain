#!/bin/sh
set -e

CONFIG_FILE="./client/docs/config.json"
SWAGGER_DIR="./tmp-swagger-gen"

# List of conflicting OperationIDs to rename (space-separated)
CONFLICTING_OPERATIONS="Params Query_Params Balance Account Code Query_BaseFee Query_Balance Query_Code Query_Account Query_DelegatorValidators Query_Proposals Query_Proposal Query_Deposits Query_Deposit Query_TallyResult Query_Votes Query_Vote Query_CommunityPool Query_UpgradedConsensusState"

echo "Finding files with conflicting OperationIDs..."

# Create backup (commented out)
# cp "$CONFIG_FILE" "${CONFIG_FILE}.backup"
# echo "Backup created: ${CONFIG_FILE}.backup"

# Process each swagger file
find "$SWAGGER_DIR" -name "*.swagger.json" | sort | while read file; do
  # Extract module path for creating unique names
  module_path=$(echo "$file" | sed 's|./tmp-swagger-gen/||' | sed 's|/query.swagger.json||' | sed 's|/service.swagger.json||')

  # Create a CamelCase prefix from the module path
  prefix=$(echo "$module_path" | awk -F'/' '{
    result = ""
    for (i = 2; i <= NF; i++) {
      word = $i
      result = result toupper(substr(word, 1, 1)) substr(word, 2)
    }
    print result
  }' | sed 's|-||g' | sed 's|_||g')

  # Find which conflicting operations exist in this file
  operations_found=""
  found_conflicts=0

  for op_id in $CONFLICTING_OPERATIONS; do
    # Check if this operationId exists in the file
    if jq -e --arg opid "$op_id" '
      .paths[][] |
      select(.operationId? == $opid)
    ' "$file" > /dev/null 2>&1; then
      new_name="${prefix}${op_id}"
      # Store as "old:new" pairs
      if [ -z "$operations_found" ]; then
        operations_found="$op_id:$new_name"
      else
        operations_found="$operations_found|$op_id:$new_name"
      fi
      found_conflicts=1
    fi
  done

  # If conflicts found, update config.json
  if [ "$found_conflicts" = "1" ]; then
    # echo "Processing: $file"
    # echo "  Prefix: $prefix"

    # Build the rename object for jq
    rename_json="{"
    first=1

    # Parse the operations_found string
    IFS='|'
    for pair in $operations_found; do
      old_name=$(echo "$pair" | cut -d':' -f1)
      new_name=$(echo "$pair" | cut -d':' -f2)
      # echo "    $old_name -> $new_name"

      if [ "$first" = "1" ]; then
        rename_json="${rename_json}\"$old_name\":\"$new_name\""
        first=0
      else
        rename_json="${rename_json},\"$old_name\":\"$new_name\""
      fi
    done
    unset IFS

    rename_json="${rename_json}}"

    # Check if this URL already exists in config
    if jq -e --arg url "$file" '.apis[] | select(.url == $url)' "$CONFIG_FILE" > /dev/null 2>&1; then
      # Update existing entry
      jq --arg url "$file" --argjson renames "$rename_json" '
        .apis |= map(
          if .url == $url then
            if .operationIds then
              .operationIds.rename = (.operationIds.rename // {}) + $renames
            else
              .operationIds = {rename: $renames}
            end
          else
            .
          end
        )
      ' "$CONFIG_FILE" > "${CONFIG_FILE}.tmp"
    else
      # Create new entry
      jq --arg url "$file" --argjson renames "$rename_json" '
        .apis += [{
          "url": $url,
          "operationIds": {
            "rename": $renames
          }
        }]
      ' "$CONFIG_FILE" > "${CONFIG_FILE}.tmp"
    fi

    mv "${CONFIG_FILE}.tmp" "$CONFIG_FILE"
  fi
done

echo "✓ Config.json updated with OperationID renames"
echo "Updated entries: $(jq '[.apis[] | select(.operationIds.rename)] | length' "$CONFIG_FILE")"
