Simple Discord bot for Rolemaster dice rolls.

Place your Discord bot API key in environment variable DISCORD_RMUBOT_TOKEN or specify it on the command line with '-t TOKEN'.
You don't need to prepend 'Bot' to this.


TODO: 
* Change input to slash commands which would allow ephemeral responses to asking user only
* Make the averages server or channel dependent so one !reset doesn't wipe the world
* When displaying averages, show everyone on the server/channel instead of just the asking user and the global total
* * This might mean storing username or displayname, or a struct containing that and the dice list
