# Solar battery monitoring

The aim of this project is to monitor an off-grid domestic battery solar setup to get regular updates
and store historical data to understand power usage and performance of the whole setup.

We'll be collecting data from a [Plasmatronics](http://www.plasmatronics.com.au/) PL80 regulator. This is a battery charger controller. So, it knows the power used by the house. It knows how charged the batteries are and how much they're being charged. It also knows the amount of power coming from the solar cells. We'll be talking to the PL80 via a PLI which is a crazy old-fashioned RS-232 serial interface.

There are a bunch of manuals which are going to be useful. We're linking to them here for easy reference:

PL80:

- [PL60/PL80 User Guide](http://www.plasmatronics.com.au/downloads/PL60.PL80.UserGuide.V6.pdf)
- [PL User Manual](http://www.plasmatronics.com.au/downloads/PLUserMan.V9.0324.pdf)
- [PL Reference Manual](http://www.plasmatronics.com.au/downloads/PL_Reference_Manual_6.3.1.pdf)

PLI:

- [PLI](http://www.plasmatronics.com.au/downloads/PLIman4.2.pdf)
- [PLI Communications](http://www.plasmatronics.com.au/downloads/PLI.Info.2.16.pdf)

Other bits of software and hardware that we might want to use:

- [Raspberry Pi](https://www.raspberrypi.org/)
- [Go](https://golang.org/)
- [Go serial port library](https://github.com/jacobsa/go-serial)
- [Python PLI code](https://github.com/jeremyvisser/pli) useful as reference.
- [InfluxDB](https://www.influxdata.com/) for storing data
- [Grafana](https://grafana.com/) for prototyping the visualisation of the data in InfluxDB
- [OpenBalena](https://www.balena.io/open/) for managing the software on the Raspberry PI using Docker.

## Environment variables

You'll need to set some environment variables for everything to work as expected. In development you can add a
`.env` file which makes thing a bit easier

- INFLUXDB_URL
- INFLUXDB_TOKEN
- INFLUXDB_BUCKET
- INFLUXDB_ORG

## Deploying to production

We're experimenting with using [deviceplane](https://deviceplane.com/) for managing the machine(s) running in "production". When an update to this repo is pushed to GitHub, a cross-platform docker image is automatically built with GitHub Actions and pushed to the Docker registry. Then, a new deploy is made with deviceplane using the new docker image and those are automatically rolled out to the necessary devices.

To make all this work some secrets need to be set up on GitHub:

- DEVICEPLANE_ACCESS_KEY
- DEVICEPLANE_PROJECT
- DOCKER_PASSWORD
- DOCKER_USERNAME
- INFLUXDB_BUCKET
- INFLUXDB_ORG
- INFLUXDB_TOKEN
- INFLUXDB_URL
