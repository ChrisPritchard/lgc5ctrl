# LG Controller

Go-based tool to control a LG (later model) TV using IP Control functionality. It probably works with other devices, though I am not sure if the encryption method is shared in other realms. Tested and used heavily on a 42" LG C5 (OLED42C54LA)

## usage

`-h` will print info. The basic needs of the tool are:

- a mac address - IF provided the tool will attempt to turn the TV on via the wake on lan feature. it will then exit - note no other settings are considered (except verbose mode) if a mac is specified
- an IP address. If not specified via `-i` it will try and load one from a `.env` file or the environment args as `TV_HOST`
- a port, the default is 9761 but can be overridden with `-p`
- a IP Control password via `-w`, that you need to gather when you enable IP Control

You can also specify a network CIDR range with `-n` - if no host is provided or the one provided doesn't connect within a 1 second timeout, the tool will attempt to scan the CIDR range for the TV.

After settings are gathered, the latest host and password are written to the local `.env` file - this is sort of a caching system that can be used with the `-n` discovery functionality; once an IP is discovered it will be tested first before the port scan is tried.

Commands are specified as a single space-joined string after the rest of the arguments. If no commands are provided, it will simply check the TV's mute state (on or off)
Some examples of what can be used:

- `VOLUME_CONTROL 10` sets volume to 10
- `INPUT_SELECT hdmi1` switches to the hdmi1 input source (this switching is why I built this, because the stupid 'smart' remote the TV ships with makes source switching a chore)

Most things can be controlled, though annoyingly not 'sound input' (e.g. switching between TV speaker and an HDMI ARC device), however for those cases you can send individual key inputs over multiple invocations (e.g. settings menu, wait, right, wait, and so on) - key commands are in the manual.

You can find all possible commands and settings in the reference docs:

- [LG_IP.pdf](./LG_IP.pdf) (also includes the encryption specification)
- [LG_RS232_IP_legacy.pdf](./LG_RS232_IP_legacy.pdf) (IP Control section)

## credits

Based on the [node implementation by Wes Souza](https://github.com/WesSouza/lgtv-ip-control). The reference docs were copied from there.
