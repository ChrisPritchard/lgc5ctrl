# LG Controller

Go-based tool to control a LG (later model) TV using IP Control functionality.

Based on the [node implementation by Wes Souza](https://github.com/WesSouza/lgtv-ip-control)

Requires TV_HOST (ip address) and TV_PASS (code setup on TV when enabling IP Control) to be set as environment variables, or in a local .env file.

Tested and used heavily on a 42" LG C5 (OLED42C54LA)

## usage

if no commands are provided, it will simply check the TV's mute state (on or off)
otherwise, all command line arguments are treated as a single space-joined string. Some examples of what can be used:

- `VOLUME_CONTROL 10` sets volume to 10
- `INPUT_SELECT hdmi1` switches to the hdmi1 input source (this switching is why i built this, because the stupid 'smart' remote they ship with makes source switching a chore)

A doc I found online with all the commands (mostly) is here: <https://www.proaudioinc.com/Dealer_Area/RS232C_EN_160526.pdf>

Though there are also some manuals in Wes' implementation.
