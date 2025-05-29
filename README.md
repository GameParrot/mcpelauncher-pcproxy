# mcpelauncher-pcproxy
Proxy intended to be used with mcpelauncher that allows you to join servers as PC. This is intended for servers that put you with mobile players when using mcpelauncher such as the hive.

# Why a proxy and not a mod
The reason this is a proxy and not a mod is because most servers kick if the Device OS does not match the Title ID. As changing the title id requires changing the auth platform, this would be very difficult in a mod. This proxy handles that by logging in using iOS preview (which has the same title ID as PC preview) so we can send a valid PC login to the server.
