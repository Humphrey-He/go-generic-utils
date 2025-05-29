#!/bin/bash

# Script to update copyright headers in Go files

# Find all Go files in the project
find . -name "*.go" -type f | while read -r file; do
  # Replace ecodeclub copyright with Humphrey
  sed -i 's|// Copyright [0-9]\{4\} ecodeclub|// Copyright 2024 Humphrey|g' "$file"
  
  # Check if file has license header
  if ! grep -q "Licensed under the Apache License" "$file"; then
    # Create a temporary file with the new header
    cat > temp_header.txt << EOL
// Copyright 2024 Humphrey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

EOL
    
    # Prepend the header to the file
    cat "$file" >> temp_header.txt
    mv temp_header.txt "$file"
  fi
  
  echo "Updated: $file"
done

# Remove the temporary file if it exists
rm -f temp_header.txt

echo "Copyright headers updated successfully" 