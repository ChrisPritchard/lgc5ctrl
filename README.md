# LG Controller

Go-based tool to control a LG (later model) TV using IP Control functionality.

Based on the [node implementation by Wes Souza](https://github.com/WesSouza/lgtv-ip-control)

Tested and used heavily on a 42" LG C5 (OLED42C54LA)

## usage

`-h` will print info. The basic needs of the tool are:

- a mac address - IF provided the tool will attempt to turn the TV on via the wake on lan feature. it will then exit - note no other settings are considered (except verbose mode) if a mac is specified
- an IP address. If not specified via `-i` it will try and load one from a `.env` file or the environment args as `TV_HOST`
- a port, the default is 9761 but can be overridden with `-p`
- a IP Control password via `-w`, that you need to gather when you enable IP Control

You can also specify a network CIDR range with `-n` - if no host is provided or the one provided doesn't connect within a 1 second timeout, the tool will attempt to scan the CIDR range for the TV.

After settings are gathered, the latest host and password are written to the local `.env` file - this is sort of a caching system that can be used with the `-n` discovery functionality.

Commands are specified as a single space joined string after the rest of the arguments. If no commands are provided, it will simply check the TV's mute state (on or off)
Some examples of what can be used:

- `VOLUME_CONTROL 10` sets volume to 10
- `INPUT_SELECT hdmi1` switches to the hdmi1 input source (this switching is why i built this, because the stupid 'smart' remote they ship with makes source switching a chore)

A doc I found online with all the commands (mostly) is here: <https://www.proaudioinc.com/Dealer_Area/RS232C_EN_160526.pdf>

Though there are also some manuals in Wes' implementation.
