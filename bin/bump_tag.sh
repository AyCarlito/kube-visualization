#!/bin/bash

# Pushes a new tag as needed to the remote.
# If the latest tag is associated with the current commit, no action is taken.
# If the latest tag is associated with a previous commit:
#   - Retrieve all commits between the current commit and commit associated with the latest tag.
#       - If any commit contains the string "BREAKING CHANGE"; increment the major, reset the minor and patch.
#       - If any commit contains the string "feat"; increment the minor and reset the patch version.
#       - if any commit contains the string "fix"; increment the patch version.
#   - The above 3 actions are mutually exclusive with one another.
#   - The new tag (if there is one) is then pushed to the remote.
#   - 