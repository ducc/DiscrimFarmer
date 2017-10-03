# DiscriminialFarmer
Changes your discord username to the username of other people to force your discriminator to change, then stops once it hits the target number(s).

![a note from the discord team](https://owo.whats-th.is/2a0e5d.png)

# Usage
1. Clone this repository
1. `$ go get`
1. `$ go build`
1. `./DiscrimFarmer -d <target discrims> -t <your discord token> -p <your discord password>`
1. Wait for a sweet discrim! Use `ctrl+c` to stop it running.

`<target discrims>` = list of discriminators separated by a comma, e.g. `1234,1337,0001`.

If you want, this tool can search for other members with the same discriminator using an api, rather than looking at the members who are on the same servers as you.
To do so, just add the flag `-api true` when running `./DiscrimFarmer`
