#!/bin/bash
# proj shell integration for bash
# Source this file in your ~/.bashrc or ~/.bash_profile

# Function to integrate proj with bash
setup_proj_bash_integration() {
    # Create proj function that handles directory changes
    proj() {
        # Store the original directory
        local original_dir="$(pwd)"
        
        # Run proj with all arguments
        command proj "$@"
        
        # Check if a .proj_last_dir file exists (created by proj on directory change)
        local proj_dir_file="$HOME/.config/proj/.proj_last_dir"
        if [ -f "$proj_dir_file" ]; then
            local target_dir="$(cat "$proj_dir_file")"
            if [ -d "$target_dir" ] && [ "$target_dir" != "$original_dir" ]; then
                echo "Changing to: $target_dir"
                cd "$target_dir"
            fi
            # Clean up the file
            rm -f "$proj_dir_file"
        fi
    }
    
    echo "proj bash integration enabled"
}

# Auto-setup if sourced
if [ "${BASH_SOURCE[0]}" != "${0}" ]; then
    setup_proj_bash_integration
fi
