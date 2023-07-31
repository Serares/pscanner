### cli app to scan ports

To add subcommands use 

`cobra add <subcommand>`

To add a sucbommand to the subcommand

`cobra add <subsubcommand> -p <subcommand[Cmd]>`

You have to use the instance variable name ending in Cmd in default cases so that the cobra cli know what associations to make