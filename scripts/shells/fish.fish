#!/usr/bin/env fish
# proj shell integration for fish
# Source this file in your ~/.config/fish/config.fish

# Function to integrate proj with fish
function setup_proj_fish_integration
    # Create proj function that handles directory changes
    function proj
        # Store the original directory
        set original_dir (pwd)
        
        # Run proj with all arguments
        command proj $argv
        
        # Check if a .proj_last_dir file exists (created by proj on directory change)
        set proj_dir_file "$HOME/.config/proj/.proj_last_dir"
        if test -f "$proj_dir_file"
            set target_dir (cat "$proj_dir_file")
            if test -d "$target_dir" -a "$target_dir" != "$original_dir"
                echo "Changing to: $target_dir"
                cd "$target_dir"
            end
            # Clean up the file
            rm -f "$proj_dir_file"
        end
    end
    
    echo "proj fish integration enabled"
end

# Auto-setup if sourced
if status --is-interactive
    setup_proj_fish_integration
end
