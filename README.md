# ipfs-alt-sync
Synchronize (and quickly resynchronize) files and directories to an unixfs hash.
 
Uses NTFS alternate data streams (ADS) to store metadata that allows future syncronization runs to skip untouched files.
These future syncronization runs can be used to bring a large directory structure up-to-date quickly* when syncronizing against a different unixfs hash that describes somewhat updated version of the same directory structure.   
 
*Existing files that should remain the same won't be blindly overwritten, but skipped instead.


## TODO
- Non-Windows support (via extended file attributes)
- Allow for the other way (dir->ipfs), so allow for quick (re)adding of slightly changed directory structure without blindly hashing the whole thing. 